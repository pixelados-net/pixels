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
	// config controls action transition timing.
	config Config
	// connections stores active transports.
	connections *netconn.Registry
	// events publishes accepted actions.
	events bus.Publisher
}

// New creates room avatar action behavior.
func New(config Config, connections *netconn.Registry, events bus.Publisher) *Service {
	return &Service{config: config.Normalize(), connections: connections, events: events}
}

// Dance cancels incompatible avatar actions and broadcasts one persistent dance.
func (service *Service) Dance(ctx context.Context, room *roomlive.Room, playerID int64, danceID int32) error {
	current, found := room.Unit(playerID)
	if !found {
		return roomlive.ErrUnitNotFound
	}
	if current.Idle {
		if err := service.setIdle(ctx, room, playerID, false, false, time.Now()); err != nil {
			return err
		}
	}
	if err := service.cancelExpression(ctx, room, current.UnitID); err != nil {
		return err
	}
	if err := service.cancelSign(ctx, room, playerID); err != nil {
		return err
	}
	if err := service.pauseTransition(ctx); err != nil {
		return err
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

// Express cancels incompatible avatar actions and broadcasts one transient expression.
func (service *Service) Express(ctx context.Context, room *roomlive.Room, playerID int64, expressionID int32) error {
	unit, found := room.Unit(playerID)
	if !found {
		return roomlive.ErrUnitNotFound
	}
	if unit.Idle {
		if err := service.setIdle(ctx, room, playerID, false, false, time.Now()); err != nil {
			return err
		}
	}
	if err := service.stopDance(ctx, room, unit); err != nil {
		return err
	}
	if err := service.cancelSign(ctx, room, playerID); err != nil {
		return err
	}
	if err := service.pauseTransition(ctx); err != nil {
		return err
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
	if idle {
		unit, found := room.Unit(playerID)
		if !found {
			return roomlive.ErrUnitNotFound
		}
		if err := service.cancelExpression(ctx, room, unit.UnitID); err != nil {
			return err
		}
		if err := service.cancelSign(ctx, room, playerID); err != nil {
			return err
		}
		if err := service.pauseTransition(ctx); err != nil {
			return err
		}
	}
	return service.setIdle(ctx, room, playerID, idle, true, time.Now())
}

// setIdleAt changes and broadcasts one AFK projection at a deterministic instant.
func (service *Service) setIdleAt(ctx context.Context, room *roomlive.Room, playerID int64, idle bool, at time.Time) error {
	return service.setIdle(ctx, room, playerID, idle, false, at)
}

// setIdle changes and broadcasts one automatic or manual AFK projection.
func (service *Service) setIdle(ctx context.Context, room *roomlive.Room, playerID int64, idle bool, manual bool, at time.Time) error {
	current, found := room.Unit(playerID)
	if !found {
		return roomlive.ErrUnitNotFound
	}
	if current.Idle == idle && (!idle || current.ManualIdle == manual) {
		return nil
	}
	dancing := hasStatus(current, worldunit.StatusDance)
	var unit roomlive.UnitSnapshot
	if manual {
		unit, found = room.SetUnitManualIdleAt(playerID, idle, at)
	} else {
		unit, found = room.SetUnitIdleAt(playerID, idle, at)
	}
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

// Sign broadcasts one transient held sign without retaining it in late-entry state.
func (service *Service) Sign(ctx context.Context, room *roomlive.Room, playerID int64, signID int32) error {
	current, found := room.Unit(playerID)
	if !found {
		return roomlive.ErrUnitNotFound
	}
	if current.Idle {
		if err := service.setIdle(ctx, room, playerID, false, false, time.Now()); err != nil {
			return err
		}
	}
	if err := service.stopDance(ctx, room, current); err != nil {
		return err
	}
	if err := service.cancelExpression(ctx, room, current.UnitID); err != nil {
		return err
	}
	if err := service.pauseTransition(ctx); err != nil {
		return err
	}
	unit, found := room.PulseUnitStatus(playerID, worldunit.StatusSign, stringValue(signID))
	if !found {
		return roomlive.ErrUnitNotFound
	}
	return broadcast.RoomUnitStatus(ctx, service.connections, room, unit, 0)
}

// ResumeForMovement clears avatar actions that cannot continue while walking.
func (service *Service) ResumeForMovement(ctx context.Context, room *roomlive.Room, previous roomlive.UnitSnapshot) error {
	if previous.Idle {
		if err := service.setIdle(ctx, room, previous.PlayerID, false, false, time.Now()); err != nil {
			return err
		}
	}
	if hasStatus(previous, worldunit.StatusDance) {
		packet, err := outedance.Encode(previous.UnitID, 0)
		if err != nil {
			return err
		}
		if err = broadcast.RoomPacket(ctx, service.connections, room, packet, 0); err != nil {
			return err
		}
	}
	return service.cancelExpression(ctx, room, previous.UnitID)
}

// Posture changes and broadcasts one free-standing posture.
func (service *Service) Posture(ctx context.Context, room *roomlive.Room, playerID int64, sitting bool) error {
	current, found := room.Unit(playerID)
	if !found {
		return roomlive.ErrUnitNotFound
	}
	if current.Idle {
		if err := service.setIdle(ctx, room, playerID, false, false, time.Now()); err != nil {
			return err
		}
	}
	if err := service.stopDance(ctx, room, current); err != nil {
		return err
	}
	if err := service.cancelExpression(ctx, room, current.UnitID); err != nil {
		return err
	}
	if err := service.cancelSign(ctx, room, playerID); err != nil {
		return err
	}
	if err := service.pauseTransition(ctx); err != nil {
		return err
	}
	unit, found := room.SetUnitPosture(playerID, sitting)
	if !found {
		return nil
	}
	return broadcast.RoomUnitStatus(ctx, service.connections, room, unit, 0)
}
