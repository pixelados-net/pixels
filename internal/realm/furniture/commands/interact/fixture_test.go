package interact

import (
	"context"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
)

// managerForTest implements furniture reads and unrelated mutations for interaction tests.
type managerForTest struct {
	// item stores the optional durable item result.
	item furnituremodel.Item
	// found reports whether the durable item exists.
	found bool
}

// FindDefinitionByID finds no definition for tests.
func (managerForTest) FindDefinitionByID(context.Context, int64) (furnituremodel.Definition, bool, error) {
	return furnituremodel.Definition{}, false, nil
}

// ListDefinitions lists no definitions for tests.
func (managerForTest) ListDefinitions(context.Context) ([]furnituremodel.Definition, error) {
	return nil, nil
}

// FindItemByID finds no durable item unless a conflict test requires one.
func (manager managerForTest) FindItemByID(context.Context, int64) (furnituremodel.Item, bool, error) {
	return manager.item, manager.found, nil
}

// ListInventory lists no inventory items for tests.
func (managerForTest) ListInventory(context.Context, int64) ([]furnituremodel.Item, error) {
	return nil, nil
}

// ListRoomItems lists no room items for tests.
func (managerForTest) ListRoomItems(context.Context, int64) ([]furnituremodel.Item, error) {
	return nil, nil
}

// Place returns no item for tests.
func (managerForTest) Place(context.Context, furnitureservice.PlaceParams) (furnituremodel.Item, error) {
	return furnituremodel.Item{}, nil
}

// Move returns no item for tests.
func (managerForTest) Move(context.Context, furnitureservice.MoveParams) (furnituremodel.Item, error) {
	return furnituremodel.Item{}, nil
}

// Pickup returns no item for tests.
func (managerForTest) Pickup(context.Context, furnitureservice.PickupParams) (furnituremodel.Item, error) {
	return furnituremodel.Item{}, nil
}
