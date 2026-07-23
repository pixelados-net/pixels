package service

import (
	"context"
	"errors"

	"github.com/niflaot/pixels/internal/realm/furniture/repository"
)

var (
	// ErrRoomBundleStoreUnavailable reports persistence without room cloning support.
	ErrRoomBundleStoreUnavailable = errors.New("furniture room bundle store unavailable")
)

// CloneRoom copies all active furniture to a new room and owner.
func (service *Service) CloneRoom(ctx context.Context, sourceRoomID int64, targetRoomID int64, targetOwnerID int64) (int, error) {
	if sourceRoomID <= 0 || targetRoomID <= 0 {
		return 0, ErrInvalidRoomID
	}
	if targetOwnerID <= 0 {
		return 0, ErrInvalidPlayerID
	}
	store, ok := service.store.(repository.RoomBundleStore)
	if !ok {
		return 0, ErrRoomBundleStoreUnavailable
	}
	return store.CloneRoomItems(ctx, sourceRoomID, targetRoomID, targetOwnerID)
}

// PreviewRoom groups template furniture without loading every item row.
func (service *Service) PreviewRoom(ctx context.Context, roomID int64) ([]RoomBundleProduct, error) {
	if roomID <= 0 {
		return nil, ErrInvalidRoomID
	}
	store, ok := service.store.(repository.RoomBundleStore)
	if !ok {
		return nil, ErrRoomBundleStoreUnavailable
	}
	stored, err := store.ListRoomBundleProducts(ctx, roomID)
	if err != nil {
		return nil, err
	}
	products := make([]RoomBundleProduct, len(stored))
	for index := range stored {
		products[index] = RoomBundleProduct(stored[index])
	}
	return products, nil
}

// roomBundleManagerAssertion verifies Service implements bundle behavior.
var roomBundleManagerAssertion RoomBundleManager = (*Service)(nil)
