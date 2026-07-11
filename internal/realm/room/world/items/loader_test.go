package furniture

import (
	"context"
	"errors"
	"testing"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// TestLoadRoomFurnitureConvertsPlacedItems verifies successful loading and conversion.
func TestLoadRoomFurnitureConvertsPlacedItems(t *testing.T) {
	manager := &fakeManager{
		definitions: []furnituremodel.Definition{chairDefinitionForTest()},
		items:       []furnituremodel.Item{placedItemForTest(2, 3, 3, 0)},
	}

	items, err := LoadRoomFurniture(context.Background(), manager, 1)
	if err != nil {
		t.Fatalf("load room furniture: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected one item, got %#v", items)
	}
	if items[0].Point.X != 3 || items[0].Point.Y != 3 || items[0].Rotation != worldunit.RotationNorth {
		t.Fatalf("unexpected item placement %#v", items[0])
	}
	if len(items[0].Definition.Slots) != 1 || items[0].Definition.Slots[0].Status != worldfurniture.SlotStatusSit {
		t.Fatalf("unexpected item slots %#v", items[0].Definition.Slots)
	}
}

// TestLoadRoomFurnitureReturnsNilWithoutItems verifies the empty-room fast path.
func TestLoadRoomFurnitureReturnsNilWithoutItems(t *testing.T) {
	items, err := LoadRoomFurniture(context.Background(), &fakeManager{}, 1)
	if err != nil {
		t.Fatalf("load room furniture: %v", err)
	}
	if items != nil {
		t.Fatalf("expected nil items, got %#v", items)
	}
}

// TestLoadRoomFurnitureSkipsItemsWithMissingDefinition verifies orphaned items are dropped.
func TestLoadRoomFurnitureSkipsItemsWithMissingDefinition(t *testing.T) {
	manager := &fakeManager{items: []furnituremodel.Item{placedItemForTest(99, 3, 3, 0)}}

	items, err := LoadRoomFurniture(context.Background(), manager, 1)
	if err != nil {
		t.Fatalf("load room furniture: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("expected orphaned item skipped, got %#v", items)
	}
}

// TestLoadRoomFurniturePropagatesStoreErrors verifies persistence errors surface.
func TestLoadRoomFurniturePropagatesStoreErrors(t *testing.T) {
	expected := errors.New("store failed")

	_, err := LoadRoomFurniture(context.Background(), &fakeManager{itemsErr: expected}, 1)
	if !errors.Is(err, expected) {
		t.Fatalf("expected items error, got %v", err)
	}

	_, err = LoadRoomFurniture(context.Background(), &fakeManager{
		items:          []furnituremodel.Item{placedItemForTest(2, 3, 3, 0)},
		definitionsErr: expected,
	}, 1)
	if !errors.Is(err, expected) {
		t.Fatalf("expected definitions error, got %v", err)
	}
}

// TestToWorldItemSkipsInventoryItems verifies unplaced items are not converted.
func TestToWorldItemSkipsInventoryItems(t *testing.T) {
	_, ok, err := ToWorldItem(furnituremodel.Item{DefinitionID: 2}, map[int64]furnituremodel.Definition{2: chairDefinitionForTest()})
	if err != nil {
		t.Fatalf("convert item: %v", err)
	}
	if ok {
		t.Fatal("expected inventory item to be skipped")
	}
}

// TestToWorldDefinitionRejectsMalformedMetadata verifies metadata parse errors surface.
func TestToWorldDefinitionRejectsMalformedMetadata(t *testing.T) {
	_, err := ToWorldDefinition(furnituremodel.Definition{Metadata: []byte("not json")})
	if err == nil {
		t.Fatal("expected malformed metadata error")
	}
}

// TestRoundHeightRoundsToNearestInteger verifies decimal height conversion.
func TestRoundHeightRoundsToNearestInteger(t *testing.T) {
	cases := map[float64]int{1.00: 1, 1.10: 1, 1.80: 2, 0: 0}
	for value, expected := range cases {
		if got := RoundHeight(value); int(got) != expected {
			t.Fatalf("round height %v: expected %d, got %d", value, expected, got)
		}
	}
}

// chairDefinitionForTest returns a one-slot sit definition for tests.
func chairDefinitionForTest() furnituremodel.Definition {
	return furnituremodel.Definition{
		Base:  sharedmodel.Base{Identity: sharedmodel.Identity{ID: 2}},
		Width: 1, Length: 1, StackHeight: 1, AllowSit: true,
		Metadata: []byte(`{"slots":[{"dx":0,"dy":0,"status":"sit","body_rotation":4}]}`),
	}
}

// placedItemForTest returns a placed furniture item fixture.
func placedItemForTest(definitionID int64, x int, y int, rotation int16) furnituremodel.Item {
	xValue, yValue, zValue := x, y, 0.0

	return furnituremodel.Item{
		Base:         sharedmodel.Base{Identity: sharedmodel.Identity{ID: 1}},
		DefinitionID: definitionID,
		X:            &xValue, Y: &yValue, Z: &zValue,
		Rotation: furnituremodel.Rotation(rotation),
	}
}

// fakeManager stubs furniture persistence for tests.
type fakeManager struct {
	definitions    []furnituremodel.Definition
	definitionsErr error
	items          []furnituremodel.Item
	itemsErr       error
}

// FindDefinitionByID finds a definition for tests.
func (manager *fakeManager) FindDefinitionByID(_ context.Context, id int64) (furnituremodel.Definition, bool, error) {
	for _, definition := range manager.definitions {
		if definition.ID == id {
			return definition, true, nil
		}
	}

	return furnituremodel.Definition{}, false, nil
}

// ListDefinitions lists definitions for tests.
func (manager *fakeManager) ListDefinitions(context.Context) ([]furnituremodel.Definition, error) {
	return manager.definitions, manager.definitionsErr
}

// FindItemByID finds an item for tests.
func (manager *fakeManager) FindItemByID(context.Context, int64) (furnituremodel.Item, bool, error) {
	return furnituremodel.Item{}, false, nil
}

// ListInventory lists inventory items for tests.
func (manager *fakeManager) ListInventory(context.Context, int64) ([]furnituremodel.Item, error) {
	return nil, nil
}

// ListRoomItems lists room items for tests.
func (manager *fakeManager) ListRoomItems(context.Context, int64) ([]furnituremodel.Item, error) {
	return manager.items, manager.itemsErr
}

// Place places an item for tests.
func (manager *fakeManager) Place(context.Context, furnitureservice.PlaceParams) (furnituremodel.Item, error) {
	return furnituremodel.Item{}, nil
}

// Move moves an item for tests.
func (manager *fakeManager) Move(context.Context, furnitureservice.MoveParams) (furnituremodel.Item, error) {
	return furnituremodel.Item{}, nil
}

// Pickup picks up an item for tests.
func (manager *fakeManager) Pickup(context.Context, furnitureservice.PickupParams) (furnituremodel.Item, error) {
	return furnituremodel.Item{}, nil
}
