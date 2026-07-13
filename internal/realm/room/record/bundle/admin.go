package bundle

import (
	"context"

	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
)

// Mark marks an active room as a bundle source.
func (service *Service) Mark(ctx context.Context, roomID int64) (roommodel.Room, error) {
	room, found, err := service.store.SetBundleTemplate(ctx, roomID, true)
	if err != nil {
		return roommodel.Room{}, err
	}
	if !found {
		return roommodel.Room{}, ErrRoomNotFound
	}
	return room, nil
}

// Unmark removes template status when no enabled offer references it.
func (service *Service) Unmark(ctx context.Context, roomID int64) (roommodel.Room, error) {
	var result roommodel.Room
	err := service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		count, err := service.store.CountActiveBundleReferences(txCtx, roomID)
		if err != nil {
			return err
		}
		if count > 0 {
			return ErrTemplateReferenced
		}
		var found bool
		result, found, err = service.store.SetBundleTemplate(txCtx, roomID, false)
		if err == nil && !found {
			return ErrRoomNotFound
		}
		return err
	})
	return result, err
}

// Templates lists marked active template rooms.
func (service *Service) Templates(ctx context.Context) ([]roommodel.Room, error) {
	return service.store.ListBundleTemplateRooms(ctx)
}

// FindTemplate validates and returns one marked template.
func (service *Service) FindTemplate(ctx context.Context, roomID int64) (roommodel.Room, bool, error) {
	room, found, err := service.rooms.FindByID(ctx, roomID)
	if err != nil || !found {
		return roommodel.Room{}, false, err
	}
	if !room.IsBundleTemplate {
		return roommodel.Room{}, false, ErrInvalidTemplate
	}
	return room, true, nil
}
