// Package action handles room avatar action commands.
package action

import (
	"context"
	"errors"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomsession "github.com/niflaot/pixels/internal/realm/room/runtime/commands/session"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	actionservice "github.com/niflaot/pixels/internal/realm/room/world/action"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// Name identifies room avatar action commands.
const Name command.Name = "room.avatar.action"

// Kind identifies one avatar action family.
type Kind uint8

const (
	// KindDance changes persistent dance state.
	KindDance Kind = iota + 1
	// KindGesture emits a transient gesture.
	KindGesture
	// KindSign emits and stores a held sign.
	KindSign
	// KindPosture changes free-standing posture.
	KindPosture
)

var (
	// ErrPlayerNotInRoom reports an action without room presence.
	ErrPlayerNotInRoom = errors.New("player not in room")
)

// Command contains one room avatar action.
type Command struct {
	// Handler stores source connection context.
	Handler netconn.Context
	// Kind identifies the action family.
	Kind Kind
	// Value stores its protocol id.
	Value int32
}

// Handler executes avatar actions.
type Handler struct {
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings stores authenticated session bindings.
	Bindings *binding.Registry
	// Runtime stores active rooms.
	Runtime *roomlive.Registry
	// Actions changes live avatar projections.
	Actions *actionservice.Service
}

// CommandName returns the command name.
func (Command) CommandName() command.Name { return Name }

// Handle executes one validated avatar action.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	player, err := roomsession.Player(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	roomID, found := player.CurrentRoom()
	if !found {
		return ErrPlayerNotInRoom
	}
	active, found := handler.Runtime.Find(roomID)
	if !found {
		return roomlive.ErrRoomNotFound
	}
	value := envelope.Command.Value
	switch envelope.Command.Kind {
	case KindDance:
		if value < 0 || value > 5 {
			return nil
		}
		return handler.Actions.Dance(ctx, active, player.ID(), value)
	case KindGesture:
		if value == 5 {
			unit, ok := active.Unit(player.ID())
			if !ok {
				return nil
			}
			return handler.Actions.SetIdle(ctx, active, player.ID(), !unit.Idle)
		}
		if !validGesture(value) {
			return nil
		}
		return handler.Actions.Express(ctx, active, player.ID(), value)
	case KindSign:
		if value < 0 || value > 18 {
			return nil
		}
		active.SetUnitStatus(player.ID(), worldunit.StatusSign, stringValue(value))
		return handler.Actions.Express(ctx, active, player.ID(), value)
	case KindPosture:
		if value != 1 && value != 2 {
			return nil
		}
		return handler.Actions.Posture(ctx, active, player.ID(), value == 1)
	default:
		return nil
	}
}

// validGesture reports supported Nitro gesture ids.
func validGesture(value int32) bool {
	return value == 1 || value == 2 || value == 3 || value == 6 || value == 7
}

// stringValue formats a small protocol id without allocation-heavy formatting.
func stringValue(value int32) string {
	if value < 10 {
		return string(rune('0' + value))
	}
	return string([]byte{byte('0' + value/10), byte('0' + value%10)})
}
