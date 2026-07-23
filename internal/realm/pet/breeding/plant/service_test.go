package plant

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
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// plantReferences supplies one immutable monsterplant fixture.
type plantReferences struct {
	// snapshot stores the fixture generation.
	snapshot *petreference.Snapshot
}

// Current returns the fixture generation.
func (references plantReferences) Current(context.Context) (*petreference.Snapshot, error) {
	return references.snapshot, nil
}

// Refresh leaves the immutable fixture unchanged.
func (references plantReferences) Refresh(context.Context) error { return nil }

// plantClock supplies one deterministic lifecycle instant.
type plantClock struct {
	// now stores the fixture time.
	now time.Time
}

// Now returns the fixture time.
func (clock plantClock) Now() time.Time { return clock.now }

// plantStore records lifecycle and terminal mutations.
type plantStore struct {
	petrecord.Store
	// pet stores the active plant.
	pet petrecord.Pet
	// consumed stores its terminal state.
	consumed string
}

// WithinTransaction runs one plant mutation.
func (store *plantStore) WithinTransaction(ctx context.Context, action func(context.Context) error) error {
	return action(ctx)
}

// Room returns the active plant.
func (store *plantStore) Room(context.Context, int64) ([]petrecord.Pet, error) {
	return []petrecord.Pet{store.pet}, nil
}

// UpdateLifecycle compare-and-swaps absolute plant deadlines.
func (store *plantStore) UpdateLifecycle(_ context.Context, petID int64, ownerID int64, growAt *time.Time, dieAt *time.Time, version int64) (petrecord.Pet, bool, error) {
	if petID != store.pet.ID || ownerID != store.pet.OwnerPlayerID || version != store.pet.Version {
		return store.pet, false, nil
	}
	store.pet.GrowAt, store.pet.DieAt = growAt, dieAt
	store.pet.Version++
	return store.pet, true, nil
}

// ConsumePlant records one terminal compare-and-swap.
func (store *plantStore) ConsumePlant(_ context.Context, petID int64, ownerID int64, _ int64, state string, version int64) (bool, error) {
	if petID != store.pet.ID || ownerID != store.pet.OwnerPlayerID || version != store.pet.Version || store.consumed != "" {
		return false, nil
	}
	store.consumed = state
	return true, nil
}

// plantRewards captures the seed furniture grant.
type plantRewards struct {
	// params stores the latest reward.
	params furnitureservice.GrantParams
	// placed stores the latest direct room placement.
	placed furnitureservice.PlaceParams
	// grants counts created furniture batches.
	grants int
}

// Grant captures one furniture reward.
func (rewards *plantRewards) Grant(_ context.Context, params furnitureservice.GrantParams) ([]furnituremodel.Item, error) {
	rewards.params = params
	rewards.grants++
	return []furnituremodel.Item{{Base: baseForPlantTest(90), OwnerPlayerID: params.OwnerPlayerID, DefinitionID: params.DefinitionID}}, nil
}

// FindDefinitionByID resolves the compost definition fixture.
func (rewards *plantRewards) FindDefinitionByID(_ context.Context, id int64) (furnituremodel.Definition, bool, error) {
	return furnituremodel.Definition{Base: baseForPlantTest(id), SpriteID: int(id), Name: "mnstr_compost", Kind: furnituremodel.KindFloor, Width: 1, Length: 1, AllowStack: true, AllowWalk: true, InteractionModesCount: 1}, id == 4830, nil
}

// ListDefinitions returns the compost definition fixture.
func (rewards *plantRewards) ListDefinitions(context.Context) ([]furnituremodel.Definition, error) {
	definition, _, _ := rewards.FindDefinitionByID(context.Background(), 4830)
	return []furnituremodel.Definition{definition}, nil
}

// Place captures one direct room furniture placement.
func (rewards *plantRewards) Place(_ context.Context, params furnitureservice.PlaceParams) (furnituremodel.Item, error) {
	rewards.placed = params
	roomID, x, y, z := params.RoomID, params.Placement.X, params.Placement.Y, params.Placement.Z
	return furnituremodel.Item{Base: baseForPlantTest(params.ItemID), DefinitionID: 4830, OwnerPlayerID: params.ActorPlayerID, RoomID: &roomID, X: &x, Y: &y, Z: &z, Rotation: params.Placement.Rotation}, nil
}

// TestSupplementUsesAbsoluteDeadlines verifies growth acceleration is deterministic.
func TestSupplementUsesAbsoluteDeadlines(t *testing.T) {
	now := time.Unix(30_000, 0)
	service, runtimeService, rooms, active, store, _ := plantFixture(t, now, now.Add(time.Hour), now.Add(48*time.Hour))
	if err := service.Supplement(context.Background(), 21, 3, 7, 1); err != nil {
		t.Fatal(err)
	}
	if store.pet.GrowAt == nil || !store.pet.GrowAt.Equal(now) || store.pet.Version != 2 {
		t.Fatalf("plant=%+v", store.pet)
	}
	current, found := runtimeService.Snapshot(21, 3)
	if !found || current.GrowAt == nil || !current.GrowAt.Equal(now) {
		t.Fatalf("runtime=%+v found=%v", current, found)
	}
	if err := service.Supplement(context.Background(), 21, 3, 8, 0); !errors.Is(err, petrecord.ErrNoRights) {
		t.Fatalf("expected owner rejection, got %v", err)
	}
	_, _, _ = rooms.Close(context.Background(), active.ID())
}

// TestHarvestConsumesOnceAndGrantsSeed verifies the terminal transaction and room removal.
func TestHarvestConsumesOnceAndGrantsSeed(t *testing.T) {
	now := time.Unix(30_000, 0)
	service, runtimeService, rooms, active, store, rewards := plantFixture(t, now, now.Add(-time.Hour), now.Add(time.Hour))
	target := plantConnection(t)
	if err := service.Harvest(context.Background(), target, 21, 3, 7); err != nil {
		t.Fatal(err)
	}
	if store.consumed != petrecord.StateHarvested || rewards.params.DefinitionID != 4582 || rewards.params.OwnerPlayerID != 7 || rewards.params.Quantity != 1 {
		t.Fatalf("state=%q reward=%+v", store.consumed, rewards.params)
	}
	if _, found := runtimeService.Snapshot(21, 3); found {
		t.Fatal("harvested plant remained active")
	}
	if err := service.Harvest(context.Background(), target, 21, 3, 7); !errors.Is(err, petrecord.ErrNoRights) {
		t.Fatalf("expected duplicate rejection, got %v", err)
	}
	_, _, _ = rooms.Close(context.Background(), active.ID())
}

// TestCompostRequiresAndConsumesDeadPlant verifies the dead lifecycle branch.
func TestCompostRequiresAndConsumesDeadPlant(t *testing.T) {
	now := time.Unix(30_000, 0)
	service, runtimeService, rooms, active, store, rewards := plantFixture(t, now, now.Add(-2*time.Hour), now.Add(-time.Hour))
	target := plantConnection(t)
	if err := service.Compost(context.Background(), target, 21, 3, 7); err != nil {
		t.Fatal(err)
	}
	if store.consumed != petrecord.StateComposted || rewards.params.DefinitionID != 4830 || rewards.placed.RoomID != 21 || rewards.placed.Placement.X != 1 || rewards.placed.Placement.Y != 0 {
		t.Fatalf("state=%q reward=%+v placement=%+v", store.consumed, rewards.params, rewards.placed)
	}
	if _, found := runtimeService.Snapshot(21, 3); found {
		t.Fatal("composted plant remained active")
	}
	compost, found := active.FurnitureItem(90)
	if !found || compost.Definition.SpriteID != 4830 || compost.Point != grid.MustPoint(1, 0) {
		t.Fatalf("compost=%+v found=%v", compost, found)
	}
	if err := service.Compost(context.Background(), target, 21, 3, 7); !errors.Is(err, petrecord.ErrNoRights) || rewards.grants != 1 {
		t.Fatalf("duplicate error=%v grants=%d", err, rewards.grants)
	}
	_, _, _ = rooms.Close(context.Background(), active.ID())
}

// plantFixture creates one loaded monsterplant room.
func plantFixture(t testing.TB, now time.Time, growAt time.Time, dieAt time.Time) (*Service, *petruntime.Service, *roomlive.Registry, *roomlive.Room, *plantStore, *plantRewards) {
	t.Helper()
	roomID, x, y, z, rotation := int64(21), 1, 0, 0.0, int16(2)
	store := &plantStore{pet: petrecord.Pet{ID: 3, OwnerPlayerID: 7, Name: "Fern", TypeID: 16, RoomID: &roomID, X: &x, Y: &y, Z: &z, Rotation: &rotation, State: petrecord.StateRoom, GrowAt: &growAt, DieAt: &dieAt, StatsAt: now, Energy: 100, Happiness: 100, Version: 1}}
	snapshot := &petreference.Snapshot{}
	snapshot.SpeciesPresent[16] = true
	snapshot.Species[16] = petrecord.Species{TypeID: 16, Plant: true, Enabled: true}
	references := plantReferences{snapshot: snapshot}
	rooms := roomlive.NewRegistry(nil)
	active, err := rooms.Activate(roomlive.Snapshot{ID: roomID, OwnerPlayerID: 7, MaxUsers: 25, AllowPets: true})
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
	config := petpolicy.Config{Enabled: true, PlantRewardDefinitionID: 4582, PlantCompostDefinitionID: 4830}
	runtimeService := petruntime.New(config, store, references, rooms, nil, nil, nil, nil, nil, nil, nil)
	runtimeService.SetClock(plantClock{now: now})
	if err = runtimeService.EnsureRoom(context.Background(), active); err != nil {
		t.Fatal(err)
	}
	rewards := &plantRewards{}
	return New(config, store, references, rewards, rooms, runtimeService, nil), runtimeService, rooms, active, store, rewards
}

// baseForPlantTest creates one durable furniture identifier.
func baseForPlantTest(id int64) sharedmodel.Base {
	return sharedmodel.Base{Identity: sharedmodel.Identity{ID: id}}
}

// plantConnection creates a live transport-backed handler context.
func plantConnection(t testing.TB) netconn.Context {
	t.Helper()
	outbound := netconn.NewHandlerRegistry()
	var target netconn.Context
	outbound.SetFallback(func(current netconn.Context, _ codec.Packet) error {
		target = current
		return nil
	}, netconn.AllowAnyActiveState(), netconn.AllowUnauthenticated())
	session, err := netconn.NewSession(netconn.SessionConfig{ID: "plant-test", Kind: "test", Outbound: outbound, Sender: func(context.Context, codec.Packet) error { return nil }, Disposer: func(context.Context, netconn.Reason) error { return nil }})
	if err != nil {
		t.Fatal(err)
	}
	if err = session.Send(context.Background(), codec.Packet{Header: 1}); err != nil {
		t.Fatal(err)
	}
	return target
}
