package search

import (
	"context"

	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
)

// favoriteRooms loads favorite room records.
func (handler Handler) favoriteRooms(ctx context.Context, playerID int64) ([]roommodel.Room, error) {
	ids, err := handler.Navigator.ListFavoriteRoomIDs(ctx, playerID)
	if err != nil {
		return nil, err
	}

	rooms, err := handler.roomsByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	rightRoomIDs := []int64{}
	if handler.RightRooms != nil {
		rightRoomIDs, err = handler.RightRooms.RoomIDsForPlayer(ctx, playerID)
		if err != nil {
			return nil, err
		}
	}
	visibleRooms := make([]roommodel.Room, 0, len(rooms))
	for _, room := range rooms {
		if favoriteVisible(playerID, room, rightRoomIDs) {
			visibleRooms = append(visibleRooms, room)
		}
	}

	return visibleRooms, nil
}

// favoriteVisible reports whether one favorite may appear to its viewer.
func favoriteVisible(playerID int64, room roommodel.Room, rightRoomIDs []int64) bool {
	if room.DoorMode != roommodel.DoorModeInvisible || room.OwnerPlayerID == playerID {
		return true
	}
	for _, roomID := range rightRoomIDs {
		if roomID == room.ID {
			return true
		}
	}
	return false
}
