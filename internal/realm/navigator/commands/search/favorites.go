package search

import (
	"context"

	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
)

// RightsChecker resolves room-scoped navigator visibility.
type RightsChecker interface {
	// HasRights reports whether a player holds explicit room rights.
	HasRights(ctx context.Context, roomID int64, playerID int64) (bool, error)
}

// favoriteRooms loads favorite room records.
func (handler Handler) favoriteRooms(ctx context.Context, playerID int64) ([]roommodel.Room, error) {
	ids, err := handler.Navigator.ListFavoriteRoomIDs(ctx, playerID)
	if err != nil {
		return nil, err
	}

	rooms := make([]roommodel.Room, 0, len(ids))
	for _, id := range ids {
		room, found, err := handler.Rooms.FindByID(ctx, id)
		if err != nil {
			return nil, err
		}
		if !found {
			continue
		}
		visible, err := handler.favoriteVisible(ctx, playerID, room)
		if err != nil {
			return nil, err
		}
		if visible {
			rooms = append(rooms, room)
		}
	}

	return rooms, nil
}

// favoriteVisible reports whether one favorite may appear to its viewer.
func (handler Handler) favoriteVisible(ctx context.Context, playerID int64, room roommodel.Room) (bool, error) {
	if room.DoorMode != roommodel.DoorModeInvisible || room.OwnerPlayerID == playerID {
		return true, nil
	}
	if handler.Rights == nil {
		return false, nil
	}

	return handler.Rights.HasRights(ctx, room.ID, playerID)
}
