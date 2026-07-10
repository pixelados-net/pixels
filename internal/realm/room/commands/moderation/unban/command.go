// Package unban ends a player's active room ban.
package unban

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/commands/control"
	roommoderation "github.com/niflaot/pixels/internal/realm/room/moderation"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
)

const (
	// Name identifies the room unban command.
	Name command.Name = "room.moderation.unban"
)

// Command unbans one room player.
type Command struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// RoomID identifies the packet room.
	RoomID int64
	// PlayerID identifies the target.
	PlayerID int64
}

// Handler handles room unban commands.
type Handler struct {
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings stores connection bindings.
	Bindings *binding.Registry
	// Moderation manages room moderation.
	Moderation roommoderation.Manager
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name { return Name }

// Handle commits an unban for the actor's current room.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	player, roomID, err := control.Actor(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	if err := control.MatchRoom(roomID, envelope.Command.RoomID); err != nil {
		return err
	}

	return handler.Moderation.Unban(ctx, roomID, player.ID(), envelope.Command.PlayerID)
}
