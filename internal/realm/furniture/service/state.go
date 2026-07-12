package service

import (
	"context"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	"github.com/niflaot/pixels/internal/realm/furniture/repository"
)

// StateParams contains one guarded furniture state mutation.
type StateParams struct {
	// ItemID identifies the placed furniture item.
	ItemID int64
	// RoomID identifies the room authorizing the mutation.
	RoomID int64
	// Expected stores the state observed by the active room.
	Expected string
	// Next stores the state to persist.
	Next string
}

// UpdateState changes a placed item's state when persistence still matches runtime.
func (service *Service) UpdateState(ctx context.Context, params StateParams) (furnituremodel.Item, error) {
	if params.ItemID <= 0 {
		return furnituremodel.Item{}, ErrInvalidItemID
	}
	if params.RoomID <= 0 {
		return furnituremodel.Item{}, ErrInvalidRoomID
	}
	updated, matched, err := service.store.UpdateItemState(ctx, repository.UpdateItemStateParams{
		ID: params.ItemID, RoomID: params.RoomID, Expected: params.Expected, Next: params.Next,
	})
	if err != nil {
		return furnituremodel.Item{}, err
	}
	if !matched {
		return furnituremodel.Item{}, ErrStateConflict
	}

	return updated, nil
}
