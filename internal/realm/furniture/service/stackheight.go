package service

import (
	"context"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	"github.com/niflaot/pixels/internal/realm/furniture/repository"
)

// UpdateStackHeight changes one placed item's exact surface override.
func (service *Service) UpdateStackHeight(ctx context.Context, itemID int64, roomID int64, heightCM *int32) (furnituremodel.Item, error) {
	if itemID <= 0 {
		return furnituremodel.Item{}, ErrInvalidItemID
	}
	if roomID <= 0 {
		return furnituremodel.Item{}, ErrInvalidRoomID
	}
	store, ok := service.store.(repository.StackHeightWriter)
	if !ok {
		return furnituremodel.Item{}, ErrStackHeightUnavailable
	}
	updated, changed, err := store.UpdateItemStackHeight(ctx, itemID, roomID, heightCM)
	if err != nil {
		return furnituremodel.Item{}, err
	}
	if !changed {
		return furnituremodel.Item{}, ErrItemNotFound
	}
	return updated, nil
}
