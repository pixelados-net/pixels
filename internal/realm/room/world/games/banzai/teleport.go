package banzai

import (
	"context"
	"time"

	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// ScheduleTeleport delays one deterministic random-teleport hop.
func ScheduleTeleport(connections *netconn.Registry, active *roomlive.Room, playerID int64, sourceID int64) error {
	teleports := active.FurnitureByInteraction("battlebanzai_random_teleport")
	destinations := teleports[:0]
	for _, item := range teleports {
		if item.ID != sourceID {
			destinations = append(destinations, item)
		}
	}
	if len(destinations) == 0 {
		return nil
	}
	frozen, err := active.SetUnitControl(playerID, worldunit.ControlFrozen)
	if err != nil {
		return err
	}
	_ = broadcast.RoomUnitStatus(context.Background(), connections, active, frozen, 0)
	destination := destinations[int((playerID+sourceID)%int64(len(destinations)))]
	active.Schedule(500*time.Millisecond, func(time.Time) {
		moved, moveErr := active.TeleportUnit(playerID, destination.Point, worldunit.RotationSouth, false, roomlive.TeleportNear)
		if moveErr == nil {
			_ = broadcast.RoomUnitStatuses(context.Background(), connections, active, []roomlive.UnitSnapshot{moved}, 0)
		}
	})
	return nil
}
