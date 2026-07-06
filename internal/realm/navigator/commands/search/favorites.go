package search

import (
	"context"

	roommodel "github.com/niflaot/pixels/internal/realm/room/model"
)

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
		if found {
			rooms = append(rooms, room)
		}
	}

	return rooms, nil
}
