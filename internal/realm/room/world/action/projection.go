package action

import (
	"context"
	"time"

	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	outdance "github.com/niflaot/pixels/networking/outbound/room/entities/dance"
	outexpression "github.com/niflaot/pixels/networking/outbound/room/entities/expression"
	"github.com/niflaot/pixels/pkg/bus"
)

// pauseTransition waits briefly between incompatible avatar projections.
func (service *Service) pauseTransition(ctx context.Context) error {
	timer := time.NewTimer(service.config.TransitionDelay)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// stopDance cancels one persistent dance when present.
func (service *Service) stopDance(ctx context.Context, room *roomlive.Room, unit roomlive.UnitSnapshot) error {
	if !hasStatus(unit, worldunit.StatusDance) {
		return nil
	}
	room.SetUnitDance(unit.PlayerID, 0)
	packet, err := outdance.Encode(unit.UnitID, 0)
	if err != nil {
		return err
	}
	return broadcast.RoomPacket(ctx, service.connections, room, packet, 0)
}

// cancelExpression clears one active client expression.
func (service *Service) cancelExpression(ctx context.Context, room *roomlive.Room, unitID int64) error {
	packet, err := outexpression.Encode(unitID, 0)
	if err != nil {
		return err
	}
	return broadcast.RoomPacket(ctx, service.connections, room, packet, 0)
}

// cancelSign clears one active client sign through a transient status projection.
func (service *Service) cancelSign(ctx context.Context, room *roomlive.Room, playerID int64) error {
	unit, found := room.PulseUnitStatus(playerID, worldunit.StatusSign, "-1")
	if !found {
		return roomlive.ErrUnitNotFound
	}
	return broadcast.RoomUnitStatus(ctx, service.connections, room, unit, 0)
}

// stringValue formats a small protocol id without general-purpose formatting.
func stringValue(value int32) string {
	if value < 10 {
		return string(rune('0' + value))
	}
	return string([]byte{byte('0' + value/10), byte('0' + value%10)})
}

// hasStatus reports whether one stable unit snapshot contains a status key.
func hasStatus(unit roomlive.UnitSnapshot, key string) bool {
	for _, status := range unit.Statuses {
		if status.Key == key {
			return true
		}
	}
	return false
}

// publish emits one optional action event.
func (service *Service) publish(ctx context.Context, name bus.Name, payload any) error {
	if service.events == nil {
		return nil
	}
	return service.events.Publish(ctx, bus.Event{Name: name, Payload: payload})
}
