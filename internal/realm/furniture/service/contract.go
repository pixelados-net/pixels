package service

import (
	"context"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
)

// DefinitionFinder reads furniture definition records.
type DefinitionFinder interface {
	// FindDefinitionByID finds a furniture definition by id.
	FindDefinitionByID(ctx context.Context, id int64) (furnituremodel.Definition, bool, error)

	// ListDefinitions lists furniture definitions.
	ListDefinitions(ctx context.Context) ([]furnituremodel.Definition, error)
}

// ItemFinder reads furniture item records.
type ItemFinder interface {
	// FindItemByID finds a furniture item by id.
	FindItemByID(ctx context.Context, id int64) (furnituremodel.Item, bool, error)

	// ListInventory lists unplaced items owned by a player.
	ListInventory(ctx context.Context, playerID int64) ([]furnituremodel.Item, error)

	// ListRoomItems lists items placed in a room.
	ListRoomItems(ctx context.Context, roomID int64) ([]furnituremodel.Item, error)
}

// Manager reads and mutates furniture persistence state.
type Manager interface {
	DefinitionFinder
	ItemFinder

	// Place moves an inventory item into a room.
	Place(ctx context.Context, params PlaceParams) (furnituremodel.Item, error)

	// Move repositions an already placed item.
	Move(ctx context.Context, params MoveParams) (furnituremodel.Item, error)

	// Pickup returns a placed item to inventory.
	Pickup(ctx context.Context, params PickupParams) (furnituremodel.Item, error)
}

// managerAssertion verifies Service implements Manager.
var managerAssertion Manager = (*Service)(nil)
