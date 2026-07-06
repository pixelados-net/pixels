// Package leave removes a player from their active room.
package leave

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/broadcast"
	roomsession "github.com/niflaot/pixels/internal/realm/room/commands/session"
	roomleft "github.com/niflaot/pixels/internal/realm/room/events/left"
	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
)

const (
	// Name identifies the room leave command.
	Name command.Name = "room.leave"
)

// Command leaves the current room.
type Command struct {
	// Handler stores the source connection handler.
	Handler netconn.Context

	// PlayerID optionally identifies the player without resolving a session.
	PlayerID int64
}

// Handler handles room leave commands.
type Handler struct {
	// Players stores live player state.
	Players *playerlive.Registry

	// Bindings stores player connection bindings.
	Bindings *binding.Registry

	// Runtime stores active rooms.
	Runtime *roomlive.Registry

	// Connections stores active network connections.
	Connections *netconn.Registry

	// Events publishes room lifecycle events.
	Events bus.Publisher
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name {
	return Name
}

// Handle handles a room leave command.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	player, playerID, err := handler.resolvePlayer(envelope.Command)
	if err != nil {
		return err
	}
	room, found := handler.Runtime.FindByPlayer(playerID)
	unitID := unitIDForPlayer(room, playerID)
	occupancy, removed, err := handler.Runtime.RemovePlayer(ctx, playerID)
	if err != nil || !removed {
		return err
	}
	if player != nil {
		player.LeaveRoom()
	}
	if found && unitID > 0 {
		_ = broadcast.RoomRemove(ctx, handler.Connections, room, unitID, playerID)
	}

	return handler.publish(ctx, roomleft.Payload{PlayerID: playerID, RoomID: occupancy.RoomID})
}

// resolvePlayer returns the leaving player identity.
func (handler Handler) resolvePlayer(roomCommand Command) (*playerlive.Player, int64, error) {
	if roomCommand.PlayerID > 0 {
		if handler.Players == nil {
			return nil, roomCommand.PlayerID, nil
		}
		player, found := handler.Players.Find(roomCommand.PlayerID)
		if !found {
			return nil, roomCommand.PlayerID, nil
		}

		return player, roomCommand.PlayerID, nil
	}

	player, err := roomsession.Player(roomCommand.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return nil, 0, err
	}

	return player, player.ID(), nil
}

// publish emits the room left event.
func (handler Handler) publish(ctx context.Context, payload roomleft.Payload) error {
	if handler.Events == nil {
		return nil
	}

	return handler.Events.Publish(ctx, bus.Event{Name: roomleft.Name, Payload: payload})
}

// unitIDForPlayer returns the live unit id for a player.
func unitIDForPlayer(room *roomlive.Room, playerID int64) int64 {
	if room == nil {
		return 0
	}
	for _, unit := range room.Units() {
		if unit.PlayerID == playerID {
			return unit.UnitID
		}
	}

	return 0
}
