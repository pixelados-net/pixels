package service

import (
	"context"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	"github.com/niflaot/pixels/internal/realm/furniture/repository"
)

// OpenGiftParams contains input for opening a placed gift.
type OpenGiftParams struct {
	// ItemID identifies the furniture item.
	ItemID int64
	// ActorPlayerID identifies the owner requesting the open.
	ActorPlayerID int64
	// RoomID identifies the room authorizing the open.
	RoomID int64
}

// OpenGift marks a placed gift as opened by its owner.
func (service *Service) OpenGift(ctx context.Context, params OpenGiftParams) (furnituremodel.Item, error) {
	if err := validateActor(params.ItemID, params.ActorPlayerID); err != nil {
		return furnituremodel.Item{}, err
	}
	if params.RoomID <= 0 {
		return furnituremodel.Item{}, ErrInvalidRoomID
	}
	item, err := service.ownedItem(ctx, params.ItemID, params.ActorPlayerID)
	if err != nil {
		return furnituremodel.Item{}, err
	}
	if item.RoomID == nil || *item.RoomID != params.RoomID {
		return furnituremodel.Item{}, ErrItemNotInRoom
	}
	if !item.GiftWrapped {
		return furnituremodel.Item{}, ErrItemNotGift
	}
	opener, ok := service.store.(repository.GiftItemOpener)
	if !ok {
		return furnituremodel.Item{}, ErrGiftOpenerUnavailable
	}

	opened, updated, err := opener.OpenGiftItem(ctx, repository.OpenGiftItemParams{
		ID: params.ItemID, OwnerPlayerID: params.ActorPlayerID, RoomID: params.RoomID,
	})
	if err != nil {
		return furnituremodel.Item{}, err
	}
	if !updated {
		return furnituremodel.Item{}, ErrItemNotGift
	}

	return opened, nil
}
