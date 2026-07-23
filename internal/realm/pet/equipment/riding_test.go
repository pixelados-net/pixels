package equipment

import (
	"context"
	"errors"
	"testing"
	"time"

	petpolicy "github.com/niflaot/pixels/internal/realm/pet/policy"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	petruntime "github.com/niflaot/pixels/internal/realm/pet/runtime"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// ridingStore persists one placed horse.
type ridingStore struct {
	petrecord.Store
	// pet stores the current aggregate.
	pet petrecord.Pet
	// updated controls optimistic writes.
	updated bool
}

// Room returns the placed horse during activation.
func (store *ridingStore) Room(_ context.Context, roomID int64) ([]petrecord.Pet, error) {
	if store.pet.RoomID == nil || *store.pet.RoomID != roomID {
		return nil, nil
	}
	return []petrecord.Pet{store.pet}, nil
}

// UpdateFlags compare-and-swaps public pet permissions.
func (store *ridingStore) UpdateFlags(_ context.Context, petID int64, ownerID int64, publicRide bool, publicBreed bool, version int64) (petrecord.Pet, bool, error) {
	if !store.updated || store.pet.ID != petID || store.pet.OwnerPlayerID != ownerID || store.pet.Version != version {
		return store.pet, false, nil
	}
	store.pet.PublicRide, store.pet.PublicBreed = publicRide, publicBreed
	store.pet.Version++
	return store.pet, true, nil
}

// SetSaddle compare-and-swaps horse equipment.
func (store *ridingStore) SetSaddle(_ context.Context, petID int64, ownerID int64, enabled bool, version int64) (petrecord.Pet, bool, error) {
	if !store.updated || store.pet.ID != petID || store.pet.OwnerPlayerID != ownerID || store.pet.Version != version {
		return store.pet, false, nil
	}
	store.pet.HasSaddle = enabled
	store.pet.Version++
	return store.pet, true, nil
}

// UpdateLifecycle compare-and-swaps monsterplant deadlines.
func (store *ridingStore) UpdateLifecycle(_ context.Context, petID int64, ownerID int64, growAt *time.Time, dieAt *time.Time, version int64) (petrecord.Pet, bool, error) {
	if !store.updated || store.pet.ID != petID || store.pet.OwnerPlayerID != ownerID || store.pet.Version != version {
		return store.pet, false, nil
	}
	store.pet.GrowAt, store.pet.DieAt = growAt, dieAt
	store.pet.Version++
	return store.pet, true, nil
}

// TestUsePlantRevivalConsumesPotion verifies the room-product revival path.
func TestUsePlantRevivalConsumesPotion(t *testing.T) {
	service, runtimeService, rooms, active, store, trading, products := productFixture(t, true)
	now := time.Now()
	growAt, dieAt := now.Add(-48*time.Hour), now.Add(-time.Hour)
	store.pet.TypeID, store.pet.CreatedAt, store.pet.GrowAt, store.pet.DieAt = 16, now.Add(-72*time.Hour), &growAt, &dieAt
	trading.item.DefinitionID, products.item.DefinitionID = 4578, 4578
	references := service.references.(productReferences).snapshot
	references.SpeciesPresent[16] = true
	references.Species[16] = petrecord.Species{TypeID: 16, Plant: true, Enabled: true}
	references.ProductRules[4578] = petrecord.ProductRule{DefinitionID: 4578, Kind: "revive", TypeID: 16, Consumable: true, Enabled: true}
	runtimeService.ReplacePlaced(store.pet)
	result, err := service.UseProduct(context.Background(), 9, 7, 99, 50)
	if err != nil {
		t.Fatal(err)
	}
	state := result.Pet.DerivePlantState(now, references.Species[16])
	if state.Dead || !state.FullyGrown || !result.Consumed || !trading.deleted || !products.picked {
		t.Fatalf("state=%+v result=%+v deleted=%v picked=%v", state, result, trading.deleted, products.picked)
	}
	_, _, _ = rooms.Close(context.Background(), active.ID())
}

// TestToggleFlagsPersistsAndRefreshesRuntime verifies both owner-managed flags.
func TestToggleFlagsPersistsAndRefreshesRuntime(t *testing.T) {
	service, runtimeService, rooms, active, store := ridingFixture(t)
	if err := service.TogglePublicRide(context.Background(), 9, 50, 7); err != nil {
		t.Fatal(err)
	}
	if err := service.TogglePublicBreed(context.Background(), 9, 50, 7); err != nil {
		t.Fatal(err)
	}
	current, found := runtimeService.Snapshot(9, 50)
	if !found || !current.PublicRide || !current.PublicBreed || current.Version != 3 {
		t.Fatalf("runtime=%+v found=%v", current, found)
	}
	if !store.pet.PublicRide || !store.pet.PublicBreed {
		t.Fatalf("store=%+v", store.pet)
	}
	_, _, _ = rooms.Close(context.Background(), active.ID())
}

// TestToggleFlagsRejectsForeignOwnerAndConflict verifies authorization and CAS failures.
func TestToggleFlagsRejectsForeignOwnerAndConflict(t *testing.T) {
	service, _, rooms, active, store := ridingFixture(t)
	if err := service.TogglePublicRide(context.Background(), 9, 50, 8); !errors.Is(err, petrecord.ErrNoRights) {
		t.Fatalf("expected owner rejection, got %v", err)
	}
	store.updated = false
	if err := service.TogglePublicRide(context.Background(), 9, 50, 7); !errors.Is(err, petrecord.ErrConflict) {
		t.Fatalf("expected optimistic conflict, got %v", err)
	}
	_, _, _ = rooms.Close(context.Background(), active.ID())
}

// TestRemoveSaddlePersistsUnequippedState verifies durable equipment removal.
func TestRemoveSaddlePersistsUnequippedState(t *testing.T) {
	service, runtimeService, rooms, active, store := ridingFixture(t)
	if err := service.RemoveSaddle(context.Background(), 9, 50, 7); err != nil {
		t.Fatal(err)
	}
	current, found := runtimeService.Snapshot(9, 50)
	if !found || current.HasSaddle || store.pet.HasSaddle || current.Version != 2 {
		t.Fatalf("runtime=%+v store=%+v found=%v", current, store.pet, found)
	}
	_, _, _ = rooms.Close(context.Background(), active.ID())
}

// TestMountAndDismountTracksRider verifies the owner riding lifecycle.
func TestMountAndDismountTracksRider(t *testing.T) {
	service, runtimeService, rooms, active, _ := ridingFixture(t)
	if _, err := active.Join(roomlive.Occupant{PlayerID: 7, Username: "demo", ConnectionID: netconn.ID("owner"), ConnectionKind: netconn.Kind("test")}); err != nil {
		t.Fatal(err)
	}
	if err := service.Mount(context.Background(), 9, 50, 7, true); err != nil {
		t.Fatal(err)
	}
	if rider, mounted := runtimeService.Rider(9, 50); !mounted || rider != 7 {
		t.Fatalf("rider=%d mounted=%v", rider, mounted)
	}
	if err := service.Mount(context.Background(), 9, 50, 7, false); err != nil {
		t.Fatal(err)
	}
	if _, mounted := runtimeService.Rider(9, 50); mounted {
		t.Fatal("rider remained mounted")
	}
	_, _, _ = rooms.Close(context.Background(), active.ID())
}

// TestGiveHandItemConsumesCarriedFood verifies hand food updates both unit and pet state.
func TestGiveHandItemConsumesCarriedFood(t *testing.T) {
	service, runtimeService, rooms, active, _ := ridingFixture(t)
	if _, err := active.Join(roomlive.Occupant{PlayerID: 7, Username: "demo", ConnectionID: netconn.ID("owner"), ConnectionKind: netconn.Kind("test")}); err != nil {
		t.Fatal(err)
	}
	if _, err := active.SetHandItem(7, 5); err != nil {
		t.Fatal(err)
	}
	unit, found := active.Unit(petruntime.EntityKey(50))
	if !found {
		t.Fatal("pet unit missing")
	}
	if err := service.GiveHandItem(context.Background(), 9, 7, unit.UnitID); err != nil {
		t.Fatal(err)
	}
	actor, found := active.Unit(7)
	pet, petFound := runtimeService.Snapshot(9, 50)
	if !found || actor.HandItem != 0 || !petFound || pet.Energy != 110 || pet.Happiness != 102 || pet.Experience != 1 {
		t.Fatalf("actor=%+v found=%v pet=%+v petFound=%v", actor, found, pet, petFound)
	}
	_, _, _ = rooms.Close(context.Background(), active.ID())
}

// ridingFixture creates one loaded saddled horse.
func ridingFixture(t testing.TB) (*Service, *petruntime.Service, *roomlive.Registry, *roomlive.Room, *ridingStore) {
	t.Helper()
	roomID, x, y, z, rotation := int64(9), 1, 0, 0.0, int16(2)
	store := &ridingStore{pet: petrecord.Pet{ID: 50, OwnerPlayerID: 7, Name: "Spirit", TypeID: 15, RoomID: &roomID, X: &x, Y: &y, Z: &z, Rotation: &rotation, State: petrecord.StateRoom, HasSaddle: true, Energy: 100, Happiness: 100, StatsAt: time.Now(), Version: 1}, updated: true}
	rooms := roomlive.NewRegistry(nil)
	active, err := rooms.Activate(roomlive.Snapshot{ID: roomID, OwnerPlayerID: 7, MaxUsers: 25, AllowPets: true, AllowPetsEat: true})
	if err != nil {
		t.Fatal(err)
	}
	roomGrid, err := grid.Parse("00", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	if err = active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}}); err != nil {
		t.Fatal(err)
	}
	config := petpolicy.Config{Enabled: true}
	runtimeService := petruntime.New(config, store, nil, rooms, nil, nil, nil, nil, nil, nil, nil)
	if err = runtimeService.EnsureRoom(context.Background(), active); err != nil {
		t.Fatal(err)
	}
	return New(config, store, nil, nil, nil, rooms, runtimeService, nil), runtimeService, rooms, active, store
}
