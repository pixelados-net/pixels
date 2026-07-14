package action

import (
	"context"
	"time"

	"github.com/niflaot/pixels/internal/realm/room/runtime/broadcast"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	roomdanced "github.com/niflaot/pixels/internal/realm/room/world/events/danced"
	roomexpressed "github.com/niflaot/pixels/internal/realm/room/world/events/expressed"
	roomidle "github.com/niflaot/pixels/internal/realm/room/world/events/idlechanged"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	netconn "github.com/niflaot/pixels/networking/connection"
	outedance "github.com/niflaot/pixels/networking/outbound/room/entities/dance"
	outeffect "github.com/niflaot/pixels/networking/outbound/room/entities/expression"
	outidle "github.com/niflaot/pixels/networking/outbound/room/entities/idle"
	"github.com/niflaot/pixels/pkg/bus"
)

// Service changes and projects live room avatar actions.
type Service struct {
	// connections stores active transports.
	connections *netconn.Registry
	// events publishes accepted actions.
	events bus.Publisher
}

// New creates room avatar action behavior.
func New(connections *netconn.Registry, events bus.Publisher) *Service {
	return &Service{connections: connections, events: events}
}

// Dance changes and broadcasts one persistent dance.
func (service *Service) Dance(ctx context.Context, room *roomlive.Room, playerID int64, danceID int32) error {
	current, found := room.Unit(playerID)
	if !found {
		return roomlive.ErrUnitNotFound
	}
	if current.Idle {
		if err := service.setIdleAt(ctx, room, playerID, false, time.Now()); err != nil {
			return err
		}
	}
	unit, found := room.SetUnitDance(playerID, danceID)
	if !found {
		return roomlive.ErrUnitNotFound
	}
	if err := broadcast.RoomUnitStatus(ctx, service.connections, room, unit, 0); err != nil {
		return err
	}
	packet, err := outedance.Encode(unit.UnitID, danceID)
	if err != nil {
		return err
	}
	if err = broadcast.RoomPacket(ctx, service.connections, room, packet, 0); err != nil {
		return err
	}
	return service.publish(ctx, roomdanced.Name, roomdanced.Payload{RoomID: room.ID(), RoomIndex: unit.UnitID, DanceID: danceID})
}

// Express broadcasts one transient expression.
func (service *Service) Express(ctx context.Context, room *roomlive.Room, playerID int64, expressionID int32) error {
	unit, found := room.Unit(playerID)
	if !found {
		return roomlive.ErrUnitNotFound
	}
	packet, err := outeffect.Encode(unit.UnitID, expressionID)
	if err != nil {
		return err
	}
	if err = broadcast.RoomPacket(ctx, service.connections, room, packet, 0); err != nil {
		return err
	}
	return service.publish(ctx, roomexpressed.Name, roomexpressed.Payload{RoomID: room.ID(), RoomIndex: unit.UnitID, ExpressionID: expressionID})
}

// SetIdle changes and broadcasts one AFK projection when needed.
func (service *Service) SetIdle(ctx context.Context, room *roomlive.Room, playerID int64, idle bool) error {
	return service.setIdleAt(ctx, room, playerID, idle, time.Now())
}

// setIdleAt changes and broadcasts one AFK projection at a deterministic instant.
func (service *Service) setIdleAt(ctx context.Context, room *roomlive.Room, playerID int64, idle bool, at time.Time) error {
	current, found := room.Unit(playerID)
	if !found {
		return roomlive.ErrUnitNotFound
	}
	if current.Idle == idle {
		return nil
	}
	dancing := hasStatus(current, worldunit.StatusDance)
	unit, found := room.SetUnitIdleAt(playerID, idle, at)
	if !found {
		return roomlive.ErrUnitNotFound
	}
	if idle && dancing {
		packet, encodeErr := outedance.Encode(unit.UnitID, 0)
		if encodeErr != nil {
			return encodeErr
		}
		if sendErr := broadcast.RoomPacket(ctx, service.connections, room, packet, 0); sendErr != nil {
			return sendErr
		}
	}
	packet, err := outidle.Encode(unit.UnitID, idle)
	if err != nil {
		return err
	}
	if err = broadcast.RoomPacket(ctx, service.connections, room, packet, 0); err != nil {
		return err
	}
	return service.publish(ctx, roomidle.Name, roomidle.Payload{RoomID: room.ID(), RoomIndex: unit.UnitID, Idle: idle})
}

// Posture changes and broadcasts one free-standing posture.
func (service *Service) Posture(ctx context.Context, room *roomlive.Room, playerID int64, sitting bool) error {
	current, found := room.Unit(playerID)
	if !found {
		return roomlive.ErrUnitNotFound
	}
	dancing := hasStatus(current, worldunit.StatusDance)
	unit, found := room.SetUnitPosture(playerID, sitting)
	if !found {
		return nil
	}
	if dancing {
		packet, err := outedance.Encode(unit.UnitID, 0)
		if err != nil {
			return err
		}
		if err = broadcast.RoomPacket(ctx, service.connections, room, packet, 0); err != nil {
			return err
		}
	}
	return broadcast.RoomUnitStatus(ctx, service.connections, room, unit, 0)
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
