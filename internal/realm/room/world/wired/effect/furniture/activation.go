package furniture

import (
	"context"
	"time"

	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	roomtask "github.com/niflaot/pixels/internal/realm/room/runtime/live/task"
	outstate "github.com/niflaot/pixels/networking/outbound/room/furniture/state"
)

const wiredActivationDuration = 500 * time.Millisecond

// Activate pulses one WIRED box without persistence or furniture events.
func (service *Service) Activate(ctx context.Context, roomID int64, itemID int64) error {
	active, found := service.rooms.Find(roomID)
	if !found {
		return nil
	}
	item, exists := active.FurnitureItem(itemID)
	if !exists {
		return nil
	}
	if item.ExtraData != "1" {
		if _, updated := active.SetFurnitureExtraData(itemID, "1"); !updated {
			return nil
		}
		if err := service.broadcastActivation(ctx, active, itemID, 1); err != nil {
			return err
		}
	}
	active.ScheduleReplacing(activationTaskKey(itemID), wiredActivationDuration, func(time.Time) {
		current, exists := active.FurnitureItem(itemID)
		if !exists || current.ExtraData != "1" {
			return
		}
		if _, updated := active.SetFurnitureExtraData(itemID, "0"); !updated {
			return
		}
		_ = service.broadcastActivation(context.Background(), active, itemID, 0)
	})
	return nil
}

// broadcastActivation sends one compact WIRED visual-state packet.
func (service *Service) broadcastActivation(ctx context.Context, active *roomlive.Room, itemID int64, value int) error {
	packet, err := outstate.Encode(itemID, value)
	if err != nil {
		return err
	}
	return broadcast.RoomPacket(ctx, service.connections, active, packet, 0)
}

// activationTaskKey creates a collision-free room task key for one WIRED box.
func activationTaskKey(itemID int64) roomtask.Key {
	return roomtask.Key(uint64(itemID)<<8 | 0xff)
}
