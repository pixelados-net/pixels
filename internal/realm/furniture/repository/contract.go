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
	CreateItems(ctx context.Context, definitionID int64, ownerPlayerID int64, quantity int32, extraData string, limitedEditionNumber *int32) ([]furnituremodel.Item, error)

	// PlaceItem moves an owned inventory item into a room.
	PlaceItem(ctx context.Context, params PlaceItemParams) (furnituremodel.Item, bool, error)

	// MoveItem repositions an owned, already placed item.
	MoveItem(ctx context.Context, params MoveItemParams) (furnituremodel.Item, bool, error)

	// PickupItem returns an owned, placed item to inventory.
	PickupItem(ctx context.Context, params PickupItemParams) (furnituremodel.Item, bool, error)

	// UpdateItemState changes one placed item's protocol-facing state with compare-and-swap.
	UpdateItemState(ctx context.Context, params UpdateItemStateParams) (furnituremodel.Item, bool, error)
}

// StackHeightWriter changes exact per-item stacking surfaces.
type StackHeightWriter interface {
	// UpdateItemStackHeight changes one placed item's exact stack surface override.
	UpdateItemStackHeight(ctx context.Context, itemID int64, roomID int64, heightCM *int32) (furnituremodel.Item, bool, error)
}

// GiftItemWriter writes wrapped furniture inventory items.
type GiftItemWriter interface {
	// CreateGiftItems creates wrapped inventory items for one recipient.
	CreateGiftItems(ctx context.Context, params GiftItemParams) ([]furnituremodel.Item, error)
}

// GiftItemOpener writes placed gift open state.
type GiftItemOpener interface {
	// OpenGiftItem marks one placed gift as opened by its owner.
	OpenGiftItem(ctx context.Context, params OpenGiftItemParams) (furnituremodel.Item, bool, error)
}

// TradingWriter reserves and transfers furniture under transaction guards.
type TradingWriter interface {
	// ReserveForMarketplace withdraws one owned inventory item into Marketplace limbo.
	ReserveForMarketplace(ctx context.Context, itemID int64, ownerPlayerID int64) (bool, error)
	// ReleaseFromMarketplace returns one reserved item to its seller inventory.
	ReleaseFromMarketplace(ctx context.Context, itemID int64, ownerPlayerID int64) (bool, error)
	// TransferFromMarketplace delivers one reserved item to its buyer.
	TransferFromMarketplace(ctx context.Context, itemID int64, sellerPlayerID int64, buyerPlayerID int64) (bool, error)
	// TransferInventoryItem transfers one unreserved inventory item between players.
	TransferInventoryItem(ctx context.Context, itemID int64, fromPlayerID int64, toPlayerID int64) (bool, error)
	// DeleteInventoryItem soft-deletes one unreserved owned inventory item.
	DeleteInventoryItem(ctx context.Context, itemID int64, ownerPlayerID int64) (bool, error)
}

// RoomBundleProduct contains a grouped room furniture definition.
type RoomBundleProduct struct {
	// DefinitionID identifies the grouped furniture definition.
	DefinitionID int64
	// Quantity stores the number of matching room items.
	Quantity int32
}

// RoomBundleStore clones and summarizes room furniture without materializing items.
type RoomBundleStore interface {
	// CloneRoomItems copies active room items to a new owner and room.
	CloneRoomItems(ctx context.Context, sourceRoomID int64, targetRoomID int64, targetOwnerID int64) (int, error)
	// ListRoomBundleProducts groups active room items by definition.
	ListRoomBundleProducts(ctx context.Context, roomID int64) ([]RoomBundleProduct, error)
}

// Store reads and writes furniture persistence records.
type Store interface {
	DefinitionReader
	ItemReader
	ItemWriter
}

// storeAssertion verifies Repository implements Store.
var storeAssertion Store = (*Repository)(nil)
