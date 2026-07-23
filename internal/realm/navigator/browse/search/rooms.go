package search

import (
	"context"

	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
)

// roomsByIDs reads rooms in one production query while retaining small test doubles.
func (handler Handler) roomsByIDs(ctx context.Context, ids []int64) ([]roommodel.Room, error) {
	if finder, ok := handler.Rooms.(roomservice.BatchFinder); ok {
		return finder.ListByIDs(ctx, ids)
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

// officialRooms reads configured official rooms from capable production stores.
func (handler Handler) officialRooms(ctx context.Context) ([]roommodel.Room, error) {
	if finder, ok := handler.Rooms.(roomservice.OfficialFinder); ok {
		return finder.ListOfficial(ctx, handler.limit())
	}
	return handler.Rooms.ListPopular(ctx, handler.limit())
}
