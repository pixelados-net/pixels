package service

import (
	"context"
	"errors"
	"testing"
)

// TestFindDefinitionByIDRejectsInvalidID verifies definition id validation.
func TestFindDefinitionByIDRejectsInvalidID(t *testing.T) {
	_, _, err := New(newFakeStore()).FindDefinitionByID(context.Background(), 0)
	if !errors.Is(err, ErrInvalidDefinitionID) {
		t.Fatalf("expected invalid definition id, got %v", err)
	}
}

// TestListDefinitionsReadsStore verifies definition listing.
func TestListDefinitionsReadsStore(t *testing.T) {
	definitions, err := New(newFakeStore()).ListDefinitions(context.Background())
	if err != nil {
		t.Fatalf("list definitions: %v", err)
	}
	if len(definitions) != 1 || definitions[0].Name != "chair_plasto" {
		t.Fatalf("unexpected definitions %#v", definitions)
	}
}

// TestFindItemByIDRejectsInvalidID verifies item id validation.
func TestFindItemByIDRejectsInvalidID(t *testing.T) {
	_, _, err := New(newFakeStore()).FindItemByID(context.Background(), 0)
	if !errors.Is(err, ErrInvalidItemID) {
		t.Fatalf("expected invalid item id, got %v", err)
	}
}

// TestListInventoryRejectsInvalidPlayer verifies inventory listing validation.
func TestListInventoryRejectsInvalidPlayer(t *testing.T) {
	_, err := New(newFakeStore()).ListInventory(context.Background(), 0)
	if !errors.Is(err, ErrInvalidPlayerID) {
		t.Fatalf("expected invalid player id, got %v", err)
	}
}

// TestListRoomItemsRejectsInvalidRoom verifies room listing validation.
func TestListRoomItemsRejectsInvalidRoom(t *testing.T) {
	_, err := New(newFakeStore()).ListRoomItems(context.Background(), 0)
	if !errors.Is(err, ErrInvalidRoomID) {
		t.Fatalf("expected invalid room id, got %v", err)
	}
}

// TestMoveRepositionsPlacedItem verifies room authorization is independent of item ownership.
func TestMoveRepositionsPlacedItem(t *testing.T) {
	store := newFakeStore()
	store.item = placedItemForTest()
	store.item.OwnerPlayerID = 99

	item, err := New(store).Move(context.Background(), MoveParams{ItemID: 1, ActorPlayerID: 7, RoomID: 1, Placement: validPlacementForTest()})
	if err != nil {
		t.Fatalf("move item: %v", err)
	}
	if !item.InRoom() || store.moveParams.RoomID != 1 {
		t.Fatalf("unexpected moved item %#v", item)
	}
}

// TestMoveRejectsInventoryItem verifies move state validation.
func TestMoveRejectsInventoryItem(t *testing.T) {
	store := newFakeStore()
	store.item = inventoryItemForTest()

	_, err := New(store).Move(context.Background(), MoveParams{ItemID: 1, ActorPlayerID: 7, RoomID: 1, Placement: validPlacementForTest()})
	if !errors.Is(err, ErrItemNotInRoom) {
		t.Fatalf("expected item outside authorized room, got %v", err)
	}
}

// TestMoveRejectsItemFromDifferentRoom verifies room guards cannot mutate foreign placements.
func TestMoveRejectsItemFromDifferentRoom(t *testing.T) {
	store := newFakeStore()
	store.item = placedItemForTest()

	_, err := New(store).Move(context.Background(), MoveParams{ItemID: 1, ActorPlayerID: 7, RoomID: 2, Placement: validPlacementForTest()})
	if !errors.Is(err, ErrItemNotInRoom) {
		t.Fatalf("expected item outside authorized room, got %v", err)
	}
}

// TestPickupRejectsInvalidInput verifies pickup input validation.
func TestPickupRejectsInvalidInput(t *testing.T) {
	_, err := New(newFakeStore()).Pickup(context.Background(), PickupParams{ActorPlayerID: 7})
	if !errors.Is(err, ErrInvalidItemID) {
		t.Fatalf("expected invalid item id, got %v", err)
	}
}

// TestGrantRejectsInvalidDefinition verifies malformed and missing definitions fail.
func TestGrantRejectsInvalidDefinition(t *testing.T) {
	_, err := New(newFakeStore()).Grant(context.Background(), GrantParams{OwnerPlayerID: 7, Quantity: 1})
	if !errors.Is(err, ErrInvalidDefinitionID) {
		t.Fatalf("expected invalid definition, got %v", err)
	}

	store := newFakeStore()
	store.found = false
	_, err = New(store).Grant(context.Background(), GrantParams{DefinitionID: 2, OwnerPlayerID: 7, Quantity: 1})
	if !errors.Is(err, ErrDefinitionNotFound) {
		t.Fatalf("expected missing definition, got %v", err)
	}
}

// TestGrantRejectsNonPositiveQuantity verifies grants require at least one item.
func TestGrantRejectsNonPositiveQuantity(t *testing.T) {
	_, err := New(newFakeStore()).Grant(context.Background(), GrantParams{DefinitionID: 2, OwnerPlayerID: 7})
	if !errors.Is(err, ErrInvalidQuantity) {
		t.Fatalf("expected invalid quantity, got %v", err)
	}
}

// TestGrantCreatesInventoryItems verifies successful grants preserve ownership.
func TestGrantCreatesInventoryItems(t *testing.T) {
	store := newFakeStore()
	items, err := New(store).Grant(context.Background(), GrantParams{
		DefinitionID: 2, OwnerPlayerID: 7, Quantity: 2, ExtraData: "1",
	})
	if err != nil {
		t.Fatalf("grant items: %v", err)
	}
	if len(items) != 2 || len(store.created) != 2 {
		t.Fatalf("unexpected granted items %#v", items)
	}
	for _, item := range items {
		if item.DefinitionID != 2 || item.OwnerPlayerID != 7 || !item.InInventory() {
			t.Fatalf("unexpected granted item %#v", item)
		}
	}
}

// TestPickupAllowsAuthorizedForeignItemAndPreservesOwner verifies room managers return items to their real owner.
func TestPickupAllowsAuthorizedForeignItemAndPreservesOwner(t *testing.T) {
	store := newFakeStore()
	store.item = placedItemForTest()
	store.pickupResult = inventoryItemForTest()

	_, err := New(store).Pickup(context.Background(), PickupParams{
		ItemID: 1, ActorPlayerID: 8, RoomID: 1, AllowForeign: true,
	})
	if err != nil {
		t.Fatalf("pickup authorized foreign item: %v", err)
	}
	if store.pickupParams.OwnerPlayerID != 7 {
		t.Fatalf("expected item to return to owner 7, got %#v", store.pickupParams)
	}
}

// TestPickupValidatesRoomAndForeignAuthority verifies pickup cannot cross room or ownership boundaries.
func TestPickupValidatesRoomAndForeignAuthority(t *testing.T) {
	store := newFakeStore()
	store.item = placedItemForTest()
	service := New(store)
	for _, test := range []struct {
		name   string
		params PickupParams
		want   error
	}{
		{name: "missing room", params: PickupParams{ItemID: 1, ActorPlayerID: 7}, want: ErrInvalidRoomID},
		{name: "wrong room", params: PickupParams{ItemID: 1, ActorPlayerID: 7, RoomID: 2}, want: ErrItemNotPlaced},
		{name: "foreign denied", params: PickupParams{ItemID: 1, ActorPlayerID: 8, RoomID: 1}, want: ErrNotItemOwner},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := service.Pickup(context.Background(), test.params)
			if !errors.Is(err, test.want) {
				t.Fatalf("expected %v, got %v", test.want, err)
			}
		})
	}
}
