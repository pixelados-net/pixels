package presence

import (
	"context"
	"errors"
	"testing"

	petpolicy "github.com/niflaot/pixels/internal/realm/pet/policy"
	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	petruntime "github.com/niflaot/pixels/internal/realm/pet/runtime"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// presenceStore persists one pet with rollback semantics.
type presenceStore struct {
	petrecord.Store
	// pet stores the only aggregate.
	pet petrecord.Pet
	// cancelled counts breeding cleanup calls.
	cancelled int
}

// WithinTransaction restores state when the callback fails.
func (store *presenceStore) WithinTransaction(ctx context.Context, operation func(context.Context) error) error {
	before := store.pet
	if err := operation(ctx); err != nil {
		store.pet = before
		return err
	}
	return nil
}

// Find returns the only live pet.
func (store *presenceStore) Find(_ context.Context, petID int64) (petrecord.Pet, bool, error) {
	return store.pet, store.pet.ID == petID && store.pet.DeletedAt == nil, nil
}

// Room returns the pet only when placed in the requested room.
func (store *presenceStore) Room(_ context.Context, roomID int64) ([]petrecord.Pet, error) {
	if store.pet.RoomID == nil || *store.pet.RoomID != roomID {
		return nil, nil
	}
	return []petrecord.Pet{store.pet}, nil
}

// CountInventory returns whether the pet is currently in inventory.
func (store *presenceStore) CountInventory(context.Context, int64) (int, error) {
	if store.pet.Inventory() {
		return 1, nil
	}
	return 0, nil
}

// Place compare-and-swaps the pet into the room.
func (store *presenceStore) Place(_ context.Context, petID int64, ownerID int64, roomID int64, x int, y int, z float64, rotation int16, version int64) (petrecord.Pet, bool, error) {
	if store.pet.ID != petID || store.pet.OwnerPlayerID != ownerID || !store.pet.Inventory() || store.pet.Version != version {
		return petrecord.Pet{}, false, nil
	}
	store.pet.RoomID, store.pet.X, store.pet.Y = &roomID, &x, &y
	store.pet.Z, store.pet.Rotation = &z, &rotation
	store.pet.State, store.pet.Version = petrecord.StateRoom, version+1
	return store.pet, true, nil
}

// Pickup compare-and-swaps the pet into inventory.
func (store *presenceStore) Pickup(_ context.Context, petID int64, roomID int64, ownerID int64, version int64) (petrecord.Pet, bool, error) {
	if store.pet.ID != petID || store.pet.OwnerPlayerID != ownerID || store.pet.RoomID == nil || *store.pet.RoomID != roomID || store.pet.Version != version {
		return petrecord.Pet{}, false, nil
	}
	store.pet.RoomID, store.pet.X, store.pet.Y = nil, nil, nil
	store.pet.Z, store.pet.Rotation = nil, nil
	store.pet.State, store.pet.Version = petrecord.StateInventory, version+1
	return store.pet, true, nil
}

// CancelBreedingPet records reservation cleanup.
func (store *presenceStore) CancelBreedingPet(context.Context, int64, int64) error {
	store.cancelled++
	return nil
}

// TestPlaceAndPickupPreserveSingleLocation verifies the full presence lifecycle.
func TestPlaceAndPickupPreserveSingleLocation(t *testing.T) {
	service, runtimeService, rooms, active, store := presenceFixture(t)
	placed, err := service.Place(context.Background(), PlaceParams{PetID: 50, ActorPlayerID: 7, RoomID: 9, Point: grid.MustPoint(1, 0)})
	if err != nil {
		t.Fatal(err)
	}
	if placed.RoomID == nil || *placed.RoomID != 9 || placed.Version != 2 {
		t.Fatalf("placed=%+v", placed)
	}
	if _, found := active.Unit(petruntime.EntityKey(50)); !found {
		t.Fatal("expected pet world unit")
	}
	if _, found := runtimeService.Snapshot(9, 50); !found {
		t.Fatal("expected pet runtime controller")
	}
	picked, err := service.Pickup(context.Background(), PickupParams{PetID: 50, ActorPlayerID: 7, RoomID: 9})
	if err != nil {
		t.Fatal(err)
	}
	if !picked.Inventory() || picked.Version != 3 || store.cancelled != 1 {
		t.Fatalf("picked=%+v cancelled=%d", picked, store.cancelled)
	}
	if _, found := active.Unit(petruntime.EntityKey(50)); found {
		t.Fatal("expected world unit removal")
	}
	if _, found := runtimeService.Snapshot(9, 50); found {
		t.Fatal("expected runtime controller removal")
	}
	_, _, _ = rooms.Close(context.Background(), active.ID())
}

// TestAdminPlacementHookRollsBackBeforeProjection verifies audit atomicity.
func TestAdminPlacementHookRollsBackBeforeProjection(t *testing.T) {
	service, runtimeService, rooms, active, store := presenceFixture(t)
	sentinel := errors.New("audit failed")
	_, err := service.PlaceAdmin(context.Background(), 50, 9, grid.MustPoint(1, 0), func(context.Context, petrecord.Pet) error { return sentinel })
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected audit failure, got %v", err)
	}
	if !store.pet.Inventory() || store.pet.Version != 1 {
		t.Fatalf("expected durable rollback, got %+v", store.pet)
	}
	if _, found := active.Unit(petruntime.EntityKey(50)); found {
		t.Fatal("expected provisional unit rollback")
	}
	if _, found := runtimeService.Snapshot(9, 50); found {
		t.Fatal("expected no projected controller")
	}
	_, _, _ = rooms.Close(context.Background(), active.ID())
}

// TestMoveSelectAndPickupRights verifies visible control and ownership boundaries.
func TestMoveSelectAndPickupRights(t *testing.T) {
	service, _, rooms, active, _ := presenceFixture(t)
	if _, err := service.Place(context.Background(), PlaceParams{PetID: 50, ActorPlayerID: 7, RoomID: 9, Point: grid.MustPoint(1, 0)}); err != nil {
		t.Fatal(err)
	}
	if err := service.Select(9, 50, 7); err != nil {
		t.Fatal(err)
	}
	if err := service.Move(context.Background(), 9, 50, 7, grid.MustPoint(2, 0)); err != nil {
		t.Fatal(err)
	}
	if err := service.Move(context.Background(), 9, 50, 8, grid.MustPoint(2, 0)); !errors.Is(err, petrecord.ErrNoRights) {
		t.Fatalf("expected move rejection, got %v", err)
	}
	if _, err := service.Pickup(context.Background(), PickupParams{PetID: 50, ActorPlayerID: 8, RoomID: 9}); !errors.Is(err, petrecord.ErrNoRights) {
		t.Fatalf("expected pickup rejection, got %v", err)
	}
	_, _, _ = rooms.Close(context.Background(), active.ID())
}

// TestPickupAdminHookRollsBackBeforeProjection verifies placed state survives failed audit.
func TestPickupAdminHookRollsBackBeforeProjection(t *testing.T) {
	service, runtimeService, rooms, active, store := presenceFixture(t)
	if _, err := service.Place(context.Background(), PlaceParams{PetID: 50, ActorPlayerID: 7, RoomID: 9, Point: grid.MustPoint(1, 0)}); err != nil {
		t.Fatal(err)
	}
	sentinel := errors.New("audit failed")
	_, err := service.PickupAdmin(context.Background(), 50, 9, func(context.Context, petrecord.Pet) error { return sentinel })
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected audit failure, got %v", err)
	}
	if store.pet.RoomID == nil || store.pet.Version != 2 {
		t.Fatalf("expected placed rollback, got %+v", store.pet)
	}
	if _, found := active.Unit(petruntime.EntityKey(50)); !found {
		t.Fatal("failed pickup removed world unit")
	}
	if _, found := runtimeService.Snapshot(9, 50); !found {
		t.Fatal("failed pickup removed runtime controller")
	}
	_, _, _ = rooms.Close(context.Background(), active.ID())
}

// TestPlacementRejectsDisabledRealmAndOccupiedTile verifies failures preserve inventory state.
func TestPlacementRejectsDisabledRealmAndOccupiedTile(t *testing.T) {
	service, _, rooms, active, store := presenceFixture(t)
	service.config.Enabled = false
	if _, err := service.Place(context.Background(), PlaceParams{PetID: 50, ActorPlayerID: 7, RoomID: 9, Point: grid.MustPoint(1, 0)}); !errors.Is(err, petrecord.ErrPetsDisabled) {
		t.Fatalf("expected disabled rejection, got %v", err)
	}
	service.config.Enabled = true
	if _, err := active.Join(roomlive.Occupant{PlayerID: 7, Username: "demo", ConnectionID: netconn.ID("presence-test"), ConnectionKind: netconn.Kind("ws")}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.Place(context.Background(), PlaceParams{PetID: 50, ActorPlayerID: 7, RoomID: 9, Point: grid.MustPoint(0, 0)}); !errors.Is(err, petrecord.ErrTileNotFree) {
		t.Fatalf("expected occupied tile rejection, got %v", err)
	}
	if !store.pet.Inventory() || store.pet.Version != 1 {
		t.Fatalf("failed placement mutated pet %+v", store.pet)
	}
	_, _, _ = rooms.Close(context.Background(), active.ID())
}

// presenceFixture creates one active flat room and inventory pet.
func presenceFixture(t testing.TB) (*Service, *petruntime.Service, *roomlive.Registry, *roomlive.Room, *presenceStore) {
	t.Helper()
	store := &presenceStore{pet: petrecord.Pet{ID: 50, OwnerPlayerID: 7, Name: "Pixel", State: petrecord.StateInventory, Energy: 100, Happiness: 100, Version: 1}}
	rooms := roomlive.NewRegistry(nil)
	active, err := rooms.Activate(roomlive.Snapshot{ID: 9, OwnerPlayerID: 7, MaxUsers: 25, AllowPets: true})
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
	config := petpolicy.Config{Enabled: true, MaxPerRoom: 5, MaxPerOwnerRoom: 5, MaxInventory: 25}
	runtimeService := petruntime.New(config, store, nil, rooms, nil, nil, nil, nil, nil, nil, nil)
	return New(config, store, rooms, nil, runtimeService, nil), runtimeService, rooms, active, store
}
