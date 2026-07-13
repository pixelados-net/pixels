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

// Granter creates new furniture inventory instances.
type Granter interface {
	// Grant creates inventory items for a player from one definition.
	Grant(ctx context.Context, params GrantParams) ([]furnituremodel.Item, error)
}

// GiftGranter creates wrapped furniture inventory instances.
type GiftGranter interface {
	// GrantGift creates wrapped inventory items for one recipient.
	GrantGift(ctx context.Context, params GiftGrantParams) ([]furnituremodel.Item, error)
}

// GiftOpener opens wrapped furniture instances.
type GiftOpener interface {
	// OpenGift marks one placed gift as opened.
	OpenGift(ctx context.Context, params OpenGiftParams) (furnituremodel.Item, error)
}

// DefinitionGranter reads definitions and creates furniture inventory instances.
type DefinitionGranter interface {
	DefinitionFinder
	Granter
}

// TeleportPairer stores durable relationships between granted teleport items.
type TeleportPairer interface {
	// PairTeleports validates and pairs two teleport items owned by one player.
	PairTeleports(ctx context.Context, ownerPlayerID int64, firstItemID int64, secondItemID int64) error
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

// StateUpdater changes durable furniture interaction state.
type StateUpdater interface {
	// UpdateState changes a placed item's protocol-facing state.
	UpdateState(ctx context.Context, params StateParams) (furnituremodel.Item, error)
}

// managerAssertion verifies Service implements Manager.
var managerAssertion Manager = (*Service)(nil)

// granterAssertion verifies Service implements Granter.
var granterAssertion Granter = (*Service)(nil)

// definitionGranterAssertion verifies Service implements DefinitionGranter.
var definitionGranterAssertion DefinitionGranter = (*Service)(nil)

// stateUpdaterAssertion verifies Service implements StateUpdater.
var stateUpdaterAssertion StateUpdater = (*Service)(nil)
