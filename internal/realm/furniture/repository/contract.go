package repository

import (
	"context"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
)

// DefinitionReader reads furniture definition records.
type DefinitionReader interface {
	// FindDefinitionByID finds an active furniture definition by id.
	FindDefinitionByID(ctx context.Context, id int64) (furnituremodel.Definition, bool, error)

	// ListDefinitions lists active furniture definitions.
	ListDefinitions(ctx context.Context) ([]furnituremodel.Definition, error)
}

// ItemReader reads furniture item records.
type ItemReader interface {
	// FindItemByID finds an active furniture item by id.
	FindItemByID(ctx context.Context, id int64) (furnituremodel.Item, bool, error)

	// ListInventoryItems lists active unplaced items owned by a player.
	ListInventoryItems(ctx context.Context, ownerPlayerID int64) ([]furnituremodel.Item, error)

	// ListRoomItems lists active items placed in a room.
	ListRoomItems(ctx context.Context, roomID int64) ([]furnituremodel.Item, error)
}

// ItemWriter writes furniture item records.
type ItemWriter interface {
	// CreateItems creates inventory items for one owner and definition.
	CreateItems(ctx context.Context, definitionID int64, ownerPlayerID int64, quantity int32, extraData string) ([]furnituremodel.Item, error)

	// PlaceItem moves an owned inventory item into a room.
	PlaceItem(ctx context.Context, params PlaceItemParams) (furnituremodel.Item, bool, error)

	// MoveItem repositions an owned, already placed item.
	MoveItem(ctx context.Context, params MoveItemParams) (furnituremodel.Item, bool, error)

	// PickupItem returns an owned, placed item to inventory.
	PickupItem(ctx context.Context, params PickupItemParams) (furnituremodel.Item, bool, error)

	// UpdateItemState changes one placed item's protocol-facing state with compare-and-swap.
	UpdateItemState(ctx context.Context, params UpdateItemStateParams) (furnituremodel.Item, bool, error)
}

// Store reads and writes furniture persistence records.
type Store interface {
	DefinitionReader
	ItemReader
	ItemWriter
}

// storeAssertion verifies Repository implements Store.
var storeAssertion Store = (*Repository)(nil)
