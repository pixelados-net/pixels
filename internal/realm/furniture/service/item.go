package service

import (
	"context"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	"github.com/niflaot/pixels/internal/realm/furniture/repository"
)

// PlaceParams contains input for placing an inventory item into a room.
type PlaceParams struct {
	// ItemID identifies the furniture item.
	ItemID int64

	// ActorPlayerID identifies the player requesting placement.
	ActorPlayerID int64

	// RoomID identifies the destination room.
	RoomID int64

	// Placement stores the destination floor coordinates and rotation.
	Placement furnituremodel.Placement

	// WallPosition stores Nitro wall coordinates for wall furniture.
	WallPosition string

	// UniqueInteractionType rejects another placed item with the same interaction.
	UniqueInteractionType string
}

// MoveParams contains input for repositioning a placed item.
type MoveParams struct {
	// ItemID identifies the furniture item.
	ItemID int64

	// ActorPlayerID identifies the player requesting the move.
	ActorPlayerID int64

	// RoomID identifies the room authorized for the move.
	RoomID int64

	// Placement stores the destination floor coordinates and rotation.
	Placement furnituremodel.Placement

	// WallPosition stores Nitro wall coordinates for wall furniture.
	WallPosition string
}

// PickupParams contains input for returning a placed item to inventory.
type PickupParams struct {
	// ItemID identifies the furniture item.
	ItemID int64

	// ActorPlayerID identifies the player requesting the pickup.
	ActorPlayerID int64

	// RoomID identifies the room authorizing the pickup.
	RoomID int64

	// AllowForeign permits an authorized room manager to pick up another player's item.
	AllowForeign bool
}

// FindItemByID finds a furniture item by id.
func (service *Service) FindItemByID(ctx context.Context, id int64) (furnituremodel.Item, bool, error) {
	if id <= 0 {
		return furnituremodel.Item{}, false, ErrInvalidItemID
	}

	return service.store.FindItemByID(ctx, id)
}

// ListInventory lists unplaced items owned by a player.
func (service *Service) ListInventory(ctx context.Context, playerID int64) ([]furnituremodel.Item, error) {
	if playerID <= 0 {
		return nil, ErrInvalidPlayerID
	}

	return service.store.ListInventoryItems(ctx, playerID)
}

// ListRoomItems lists items placed in a room.
func (service *Service) ListRoomItems(ctx context.Context, roomID int64) ([]furnituremodel.Item, error) {
	if roomID <= 0 {
		return nil, ErrInvalidRoomID
	}

	return service.store.ListRoomItems(ctx, roomID)
}

// Place moves an inventory item into a room.
func (service *Service) Place(ctx context.Context, params PlaceParams) (furnituremodel.Item, error) {
	if err := validateActor(params.ItemID, params.ActorPlayerID); err != nil {
		return furnituremodel.Item{}, err
	}
	if params.RoomID <= 0 {
		return furnituremodel.Item{}, ErrInvalidRoomID
	}
	if params.WallPosition == "" {
		if err := validatePlacement(params.Placement); err != nil {
			return furnituremodel.Item{}, err
		}
	} else if !furnituremodel.ValidWallPosition(params.WallPosition) {
		return furnituremodel.Item{}, ErrInvalidPlacement
	}

	item, err := service.ownedItem(ctx, params.ItemID, params.ActorPlayerID)
	if err != nil {
		return furnituremodel.Item{}, err
	}
	if item.InRoom() || item.MarketplaceReserved {
		return furnituremodel.Item{}, ErrItemNotInInventory
	}
	if service.staged != nil && service.staged.Contains(params.ItemID) {
		return furnituremodel.Item{}, ErrItemStaged
	}

	placed, updated, err := service.store.PlaceItem(ctx, repository.PlaceItemParams{
		ID:                    params.ItemID,
		OwnerPlayerID:         params.ActorPlayerID,
		RoomID:                params.RoomID,
		Placement:             params.Placement,
		WallPosition:          params.WallPosition,
		UniqueInteractionType: params.UniqueInteractionType,
	})
	if err != nil {
		return furnituremodel.Item{}, err
	}
	if !updated {
		return furnituremodel.Item{}, ErrItemNotInInventory
	}

	return placed, nil
}

// Move repositions an already placed item.
func (service *Service) Move(ctx context.Context, params MoveParams) (furnituremodel.Item, error) {
	if err := validateActor(params.ItemID, params.ActorPlayerID); err != nil {
		return furnituremodel.Item{}, err
	}
	if params.RoomID <= 0 {
		return furnituremodel.Item{}, ErrInvalidRoomID
	}
	if params.WallPosition == "" {
		if err := validatePlacement(params.Placement); err != nil {
			return furnituremodel.Item{}, err
		}
	} else if !furnituremodel.ValidWallPosition(params.WallPosition) {
		return furnituremodel.Item{}, ErrInvalidPlacement
	}

	item, found, err := service.store.FindItemByID(ctx, params.ItemID)
	if err != nil {
		return furnituremodel.Item{}, err
	}
	if !found {
		return furnituremodel.Item{}, ErrItemNotFound
	}
	if item.RoomID == nil || *item.RoomID != params.RoomID {
		return furnituremodel.Item{}, ErrItemNotInRoom
	}

	moved, updated, err := service.store.MoveItem(ctx, repository.MoveItemParams{
		ID:           params.ItemID,
		RoomID:       params.RoomID,
		Placement:    params.Placement,
		WallPosition: params.WallPosition,
	})
	if err != nil {
		return furnituremodel.Item{}, err
	}
	if !updated {
		return furnituremodel.Item{}, ErrItemNotInRoom
	}

	return moved, nil
}

// Pickup returns a placed item to inventory.
func (service *Service) Pickup(ctx context.Context, params PickupParams) (furnituremodel.Item, error) {
	if err := validateActor(params.ItemID, params.ActorPlayerID); err != nil {
		return furnituremodel.Item{}, err
	}
	if params.RoomID <= 0 {
		return furnituremodel.Item{}, ErrInvalidRoomID
	}
	item, found, err := service.store.FindItemByID(ctx, params.ItemID)
	if err != nil {
		return furnituremodel.Item{}, err
	}
	if !found {
		return furnituremodel.Item{}, ErrItemNotFound
	}
	if item.OwnerPlayerID != params.ActorPlayerID && !params.AllowForeign {
		return furnituremodel.Item{}, ErrNotItemOwner
	}
	if item.RoomID == nil || *item.RoomID != params.RoomID {
		return furnituremodel.Item{}, ErrItemNotPlaced
	}

	picked, updated, err := service.store.PickupItem(ctx, repository.PickupItemParams{
		ID:            params.ItemID,
		OwnerPlayerID: item.OwnerPlayerID,
	})
	if err != nil {
		return furnituremodel.Item{}, err
	}
	if !updated {
		return furnituremodel.Item{}, ErrItemNotPlaced
	}

	return picked, nil
}

// ownedItem finds an item and verifies the actor owns it.
func (service *Service) ownedItem(ctx context.Context, itemID int64, actorPlayerID int64) (furnituremodel.Item, error) {
	item, found, err := service.store.FindItemByID(ctx, itemID)
	if err != nil {
		return furnituremodel.Item{}, err
	}
	if !found {
		return furnituremodel.Item{}, ErrItemNotFound
	}
	if item.OwnerPlayerID != actorPlayerID {
		return furnituremodel.Item{}, ErrNotItemOwner
	}

	return item, nil
}
