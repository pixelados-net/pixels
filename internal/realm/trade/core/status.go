package core

import (
	"context"

	roombroadcast "github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
)

// projectStatuses broadcasts current direct-trade unit states to the room.
func (service *Service) projectStatuses(ctx context.Context, room *roomlive.Room, playerIDs ...int64) {
	if service.connections == nil {
		return
	}
	units := make([]roomlive.UnitSnapshot, 0, len(playerIDs))
	for _, playerID := range playerIDs {
		if unit, found := room.Unit(playerID); found {
			units = append(units, unit)
		}
	}
	_ = roombroadcast.RoomUnitStatuses(ctx, service.connections, room, units, 0)
}
