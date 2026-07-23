package equipment

import (
	"context"
	"errors"
	"testing"
	"time"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	petpolicy "github.com/niflaot/pixels/internal/realm/pet/policy"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	petreference "github.com/niflaot/pixels/internal/realm/pet/reference"
	petruntime "github.com/niflaot/pixels/internal/realm/pet/runtime"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	outrefresh "github.com/niflaot/pixels/networking/outbound/inventory/furniture/refresh"
	outremove "github.com/niflaot/pixels/networking/outbound/inventory/furniture/remove"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// productReferences supplies one immutable product generation.
type productReferences struct {
	// snapshot stores the fixture generation.
	snapshot *petreference.Snapshot
}

// Current returns the fixture generation.
func (references productReferences) Current(context.Context) (*petreference.Snapshot, error) {
	return references.snapshot, nil
}

// Refresh leaves the immutable fixture unchanged.
func (references productReferences) Refresh(context.Context) error { return nil }

// productTrading supplies and consumes one furniture product.
type productTrading struct {
	furnitureservice.TradingManager
	// item stores the placed product.
	item furnituremodel.Item
	// deleted reports committed consumption.
	deleted bool
}

// FindItemByID returns the fixture product.
func (trading *productTrading) FindItemByID(_ context.Context, id int64) (furnituremodel.Item, bool, error) {
	return trading.item, trading.item.ID == id, nil
}

// DeleteInventoryItem records committed consumption.
func (trading *productTrading) DeleteInventoryItem(_ context.Context, itemID int64, ownerID int64) error {
	if trading.item.ID != itemID || trading.item.OwnerPlayerID != ownerID {
		return errors.New("unexpected product deletion")
	}
	trading.deleted = true
	return nil
}

// productRooms records the transactional room pickup.
type productRooms struct {
	furnitureservice.Manager
	// item stores the placed product.
	item furnituremodel.Item
	// picked reports a successful pickup.
	picked bool
}

// Pickup records the product pickup.
func (rooms *productRooms) Pickup(_ context.Context, params furnitureservice.PickupParams) (furnituremodel.Item, error) {
	if params.ItemID != rooms.item.ID || params.RoomID != *rooms.item.RoomID {
		return furnituremodel.Item{}, errors.New("unexpected product pickup")
	}
	rooms.picked = true
	return rooms.item, nil
}

// WithinTransaction runs one product mutation.
func (store *ridingStore) WithinTransaction(ctx context.Context, action func(context.Context) error) error {
	return action(ctx)
}

// UpdateStats applies bounded fixture stat deltas.
func (store *ridingStore) UpdateStats(_ context.Context, petID int64, energy int32, happiness int32, experience int32, version int64) (petrecord.Pet, bool, error) {
	if !store.updated || store.pet.ID != petID || store.pet.Version != version {
		return store.pet, false, nil
	}
	store.pet.Energy += energy
	store.pet.Happiness += happiness
	store.pet.Experience += experience
	store.pet.Version++
	return store.pet, true, nil
}

// TestUseProductConsumesPlacedFoodAtomically verifies the complete room product lifecycle.
func TestUseProductConsumesPlacedFoodAtomically(t *testing.T) {
	service, runtimeService, rooms, active, store, trading, products := productFixture(t, true)
	result, err := service.UseProduct(context.Background(), 9, 7, 99, 50)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Consumed || result.Pet.Energy != 30 || result.Pet.Happiness != 25 || result.Pet.Experience != 3 {
		t.Fatalf("result=%+v", result)
	}
	if !trading.deleted || !products.picked {
		t.Fatalf("deleted=%v picked=%v", trading.deleted, products.picked)
	}
	if _, found := active.FurnitureItem(99); found {
		t.Fatal("consumed product remained in active room")
	}
	current, found := runtimeService.Snapshot(9, 50)
	if !found || current.Version != 2 || store.pet.Version != 2 {
		t.Fatalf("runtime=%+v store=%+v", current, store.pet)
	}
	_, _, _ = rooms.Close(context.Background(), active.ID())
}

// TestUseProductHonorsRoomFeedingPolicy verifies food cannot bypass room settings.
func TestUseProductHonorsRoomFeedingPolicy(t *testing.T) {
	service, _, rooms, active, store, trading, products := productFixture(t, false)
	_, err := service.UseProduct(context.Background(), 9, 7, 99, 50)
	if !errors.Is(err, petrecord.ErrFeedingDisabled) {
		t.Fatalf("expected feeding rejection, got %v", err)
	}
	if trading.deleted || products.picked || store.pet.Version != 1 {
		t.Fatalf("unexpected mutation deleted=%v picked=%v pet=%+v", trading.deleted, products.picked, store.pet)
	}
	_, _, _ = rooms.Close(context.Background(), active.ID())
}

// TestConsumeNeedRevalidatesAndConsumesTarget verifies autonomous use shares the transaction boundary.
func TestConsumeNeedRevalidatesAndConsumesTarget(t *testing.T) {
	service, runtimeService, rooms, active, _, trading, products := productFixture(t, true)
	if err := service.ConsumeNeed(context.Background(), 9, 50, 99); !errors.Is(err, petrecord.ErrInvalidProduct) {
		t.Fatalf("expected remote consumption rejection, got %v", err)
	}
	if _, err := active.TeleportUnit(petruntime.EntityKey(50), grid.MustPoint(2, 0), worldunit.RotationSouth, false); err != nil {
		t.Fatal(err)
	}
	if err := service.ConsumeNeed(context.Background(), 9, 50, 99); err != nil {
		t.Fatal(err)
	}
	current, found := runtimeService.Snapshot(9, 50)
	if !found || current.Energy != 30 || current.Happiness != 25 || !trading.deleted || !products.picked {
		t.Fatalf("pet=%+v found=%v deleted=%v picked=%v", current, found, trading.deleted, products.picked)
	}
	_, _, _ = rooms.Close(context.Background(), active.ID())
}

// TestSendProductInventoryChangeSendsRemoveThenRefresh verifies post-commit packet ordering.
func TestSendProductInventoryChangeSendsRemoveThenRefresh(t *testing.T) {
	target, packets := productConnection(t)
	service := &Service{}
	if err := service.SendProductInventoryChange(context.Background(), target, 99, true); err != nil {
		t.Fatal(err)
	}
	if len(*packets) != 2 || (*packets)[0].Header != outremove.Header || (*packets)[1].Header != outrefresh.Header {
		t.Fatalf("packets=%+v", *packets)
	}
	*packets = (*packets)[:0]
	if err := service.SendProductInventoryChange(context.Background(), target, 99, false); err != nil || len(*packets) != 0 {
		t.Fatalf("non-consumable err=%v packets=%+v", err, *packets)
	}
}

// productFixture creates one placed pet and typed food furniture.
func productFixture(t testing.TB, allowFeeding bool) (*Service, *petruntime.Service, *roomlive.Registry, *roomlive.Room, *ridingStore, *productTrading, *productRooms) {
	t.Helper()
	roomID, x, y, z, rotation := int64(9), 1, 0, 0.0, int16(2)
	store := &ridingStore{pet: petrecord.Pet{ID: 50, OwnerPlayerID: 7, Name: "Pixel", TypeID: 0, RoomID: &roomID, X: &x, Y: &y, Z: &z, Rotation: &rotation, State: petrecord.StateRoom, Energy: 20, Happiness: 20, StatsAt: time.Now(), Version: 1}, updated: true}
	rooms := roomlive.NewRegistry(nil)
	active, err := rooms.Activate(roomlive.Snapshot{ID: roomID, OwnerPlayerID: 7, MaxUsers: 25, AllowPets: true, AllowPetsEat: allowFeeding})
	if err != nil {
		t.Fatal(err)
	}
	roomGrid, err := grid.Parse("000", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if err = active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}}); err != nil {
		t.Fatal(err)
	}
	item := furnituremodel.Item{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 99}}, DefinitionID: 1532, OwnerPlayerID: 7, RoomID: &roomID, X: &x, Y: &y, Z: &z}
	worldItem := worldfurniture.Item{ID: item.ID, OwnerPlayerID: 7, Definition: worldfurniture.Definition{Width: 1, Length: 1, AllowWalk: true}, Point: grid.MustPoint(2, 0)}
	if _, err = active.ReloadFurniture(item.ID, &worldItem); err != nil {
		t.Fatal(err)
	}
	references := productReferences{snapshot: &petreference.Snapshot{ProductRules: map[int64]petrecord.ProductRule{1532: {DefinitionID: 1532, Kind: "food", TypeID: -1, EnergyDelta: 10, HappinessDelta: 5, ExperienceDelta: 3, Consumable: true, Enabled: true}}}}
	config := petpolicy.Config{Enabled: true}
	runtimeService := petruntime.New(config, store, references, rooms, nil, nil, nil, nil, nil, nil, nil)
	if err = runtimeService.EnsureRoom(context.Background(), active); err != nil {
		t.Fatal(err)
	}
	trading := &productTrading{item: item}
	products := &productRooms{item: item}
	return New(config, store, references, trading, products, rooms, runtimeService, nil), runtimeService, rooms, active, store, trading, products
}

// productConnection creates a transport-backed handler context and packet sink.
func productConnection(t testing.TB) (netconn.Context, *[]codec.Packet) {
	t.Helper()
	outbound := netconn.NewHandlerRegistry()
	var target netconn.Context
	packets := make([]codec.Packet, 0, 2)
	outbound.SetFallback(func(current netconn.Context, _ codec.Packet) error {
		target = current
		return nil
	}, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	session, err := netconn.NewSession(netconn.SessionConfig{ID: "product-test", Kind: "test", Outbound: outbound, Sender: func(_ context.Context, packet codec.Packet) error {
		packets = append(packets, packet)
		return nil
	}, Disposer: func(context.Context, netconn.Reason) error { return nil }})
	if err != nil {
		t.Fatal(err)
	}
	if err = session.Send(context.Background(), codec.Packet{Header: 1}); err != nil {
		t.Fatal(err)
	}
	packets = packets[:0]
	return target, &packets
}
