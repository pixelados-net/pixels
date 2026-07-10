package service

import (
	"context"
	"errors"
	"testing"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	"github.com/niflaot/pixels/internal/realm/furniture/repository"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// TestPlaceMovesInventoryItemIntoRoom verifies successful placement.
func TestPlaceMovesInventoryItemIntoRoom(t *testing.T) {
	store := newFakeStore()
	store.item = inventoryItemForTest()

	item, err := New(store).Place(context.Background(), validPlaceForTest())
	if err != nil {
		t.Fatalf("place item: %v", err)
	}
	if !item.InRoom() {
		t.Fatalf("expected placed item %#v", item)
	}
}

// TestPlaceRejectsInvalidInput verifies placement input validation.
func TestPlaceRejectsInvalidInput(t *testing.T) {
	cases := []struct {
		name     string
		params   PlaceParams
		expected error
	}{
		{name: "item id", params: PlaceParams{ActorPlayerID: 7, RoomID: 1, Placement: validPlacementForTest()}, expected: ErrInvalidItemID},
		{name: "actor id", params: PlaceParams{ItemID: 1, RoomID: 1, Placement: validPlacementForTest()}, expected: ErrInvalidPlayerID},
		{name: "room id", params: PlaceParams{ItemID: 1, ActorPlayerID: 7, Placement: validPlacementForTest()}, expected: ErrInvalidRoomID},
		{name: "rotation", params: PlaceParams{ItemID: 1, ActorPlayerID: 7, RoomID: 1, Placement: furnituremodel.Placement{Rotation: 1}}, expected: ErrInvalidPlacement},
		{name: "negative x", params: PlaceParams{ItemID: 1, ActorPlayerID: 7, RoomID: 1, Placement: furnituremodel.Placement{X: -1, Rotation: furnituremodel.RotationNorth}}, expected: ErrInvalidPlacement},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			_, err := New(newFakeStore()).Place(context.Background(), test.params)
			if !errors.Is(err, test.expected) {
				t.Fatalf("expected %v, got %v", test.expected, err)
			}
		})
	}
}

// TestPlaceRejectsMissingItem verifies placement against a missing item.
func TestPlaceRejectsMissingItem(t *testing.T) {
	store := newFakeStore()
	store.found = false

	_, err := New(store).Place(context.Background(), validPlaceForTest())
	if !errors.Is(err, ErrItemNotFound) {
		t.Fatalf("expected item not found, got %v", err)
	}
}

// TestPlaceRejectsNonOwner verifies placement ownership validation.
func TestPlaceRejectsNonOwner(t *testing.T) {
	store := newFakeStore()
	store.item = inventoryItemForTest()
	store.item.OwnerPlayerID = 99

	_, err := New(store).Place(context.Background(), validPlaceForTest())
	if !errors.Is(err, ErrNotItemOwner) {
		t.Fatalf("expected not item owner, got %v", err)
	}
}

// TestPlaceRejectsAlreadyPlacedItem verifies placement inventory-state validation.
func TestPlaceRejectsAlreadyPlacedItem(t *testing.T) {
	store := newFakeStore()
	store.item = placedItemForTest()

	_, err := New(store).Place(context.Background(), validPlaceForTest())
	if !errors.Is(err, ErrItemNotInInventory) {
		t.Fatalf("expected item not in inventory, got %v", err)
	}
}

// TestPlaceRejectsConcurrentConflict verifies placement race handling.
func TestPlaceRejectsConcurrentConflict(t *testing.T) {
	store := newFakeStore()
	store.item = inventoryItemForTest()
	store.placeUpdated = false

	_, err := New(store).Place(context.Background(), validPlaceForTest())
	if !errors.Is(err, ErrItemNotInInventory) {
		t.Fatalf("expected conflict mapped to inventory error, got %v", err)
	}
}

// TestPickupReturnsPlacedItemToInventory verifies successful pickup.
func TestPickupReturnsPlacedItemToInventory(t *testing.T) {
	store := newFakeStore()
	store.item = placedItemForTest()
	store.pickupResult = inventoryItemForTest()

	item, err := New(store).Pickup(context.Background(), PickupParams{ItemID: 1, ActorPlayerID: 7})
	if err != nil {
		t.Fatalf("pickup item: %v", err)
	}
	if !item.InInventory() {
		t.Fatalf("unexpected picked up item %#v", item)
	}
}

// TestPickupRejectsInventoryItem verifies pickup state validation.
func TestPickupRejectsInventoryItem(t *testing.T) {
	store := newFakeStore()
	store.item = inventoryItemForTest()

	_, err := New(store).Pickup(context.Background(), PickupParams{ItemID: 1, ActorPlayerID: 7})
	if !errors.Is(err, ErrItemNotPlaced) {
		t.Fatalf("expected item not placed, got %v", err)
	}
}

// validPlaceForTest returns valid placement input.
func validPlaceForTest() PlaceParams {
	return PlaceParams{ItemID: 1, ActorPlayerID: 7, RoomID: 1, Placement: validPlacementForTest()}
}

// validPlacementForTest returns a valid floor placement.
func validPlacementForTest() furnituremodel.Placement {
	return furnituremodel.Placement{X: 4, Y: 4, Z: 0, Rotation: furnituremodel.RotationNorth}
}

// inventoryItemForTest returns an unplaced item fixture.
func inventoryItemForTest() furnituremodel.Item {
	return furnituremodel.Item{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 1}}, OwnerPlayerID: 7}
}

// placedItemForTest returns a placed item fixture.
func placedItemForTest() furnituremodel.Item {
	roomID := int64(1)
	x, y := 4, 4
	z := 0.0
	return furnituremodel.Item{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 1}}, OwnerPlayerID: 7, RoomID: &roomID, X: &x, Y: &y, Z: &z}
}

// newFakeStore creates a furniture store for tests.
func newFakeStore() *fakeStore {
	return &fakeStore{found: true, placeUpdated: true, moveUpdated: true, pickupUpdated: true}
}

// fakeStore records furniture store calls for tests.
type fakeStore struct {
	// item is the returned item for finds and successful mutations.
	item furnituremodel.Item
	// found reports whether FindItemByID succeeds.
	found bool
	// placeUpdated reports whether PlaceItem matched a row.
	placeUpdated bool
	// moveUpdated reports whether MoveItem matched a row.
	moveUpdated bool
	// moveParams stores the latest move mutation input.
	moveParams repository.MoveItemParams
	// pickupUpdated reports whether PickupItem matched a row.
	pickupUpdated bool
	// pickupResult is the returned item for a successful pickup.
	pickupResult furnituremodel.Item
	// created stores items returned by CreateItems.
	created []furnituremodel.Item
}

// FindDefinitionByID finds a definition for tests.
func (store *fakeStore) FindDefinitionByID(context.Context, int64) (furnituremodel.Definition, bool, error) {
	return furnituremodel.Definition{Name: "chair_plasto"}, store.found, nil
}

// ListDefinitions lists definitions for tests.
func (store *fakeStore) ListDefinitions(context.Context) ([]furnituremodel.Definition, error) {
	return []furnituremodel.Definition{{Name: "chair_plasto"}}, nil
}

// FindItemByID finds an item for tests.
func (store *fakeStore) FindItemByID(context.Context, int64) (furnituremodel.Item, bool, error) {
	return store.item, store.found, nil
}

// ListInventoryItems lists inventory items for tests.
func (store *fakeStore) ListInventoryItems(context.Context, int64) ([]furnituremodel.Item, error) {
	return []furnituremodel.Item{store.item}, nil
}

// ListRoomItems lists room items for tests.
func (store *fakeStore) ListRoomItems(context.Context, int64) ([]furnituremodel.Item, error) {
	return []furnituremodel.Item{store.item}, nil
}

// CreateItems creates inventory items for tests.
func (store *fakeStore) CreateItems(_ context.Context, definitionID int64, ownerPlayerID int64, quantity int32, _ string) ([]furnituremodel.Item, error) {
	items := make([]furnituremodel.Item, 0, quantity)
	for index := int32(0); index < quantity; index++ {
		items = append(items, furnituremodel.Item{DefinitionID: definitionID, OwnerPlayerID: ownerPlayerID})
	}
	store.created = items

	return items, nil
}

// PlaceItem places an item for tests.
func (store *fakeStore) PlaceItem(_ context.Context, params repository.PlaceItemParams) (furnituremodel.Item, bool, error) {
	roomID := params.RoomID
	item := store.item
	item.RoomID = &roomID

	return item, store.placeUpdated, nil
}

// MoveItem moves an item for tests.
func (store *fakeStore) MoveItem(_ context.Context, params repository.MoveItemParams) (furnituremodel.Item, bool, error) {
	store.moveParams = params

	return store.item, store.moveUpdated, nil
}

// PickupItem picks up an item for tests.
func (store *fakeStore) PickupItem(context.Context, repository.PickupItemParams) (furnituremodel.Item, bool, error) {
	return store.pickupResult, store.pickupUpdated, nil
}
