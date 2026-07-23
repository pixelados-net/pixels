package moderation

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/control/commands/resolve"
	roommoderation "github.com/niflaot/pixels/internal/realm/room/control/moderation"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
)

const (
	// UnbanName identifies the room unban command.
	UnbanName command.Name = "room.moderation.unban"
)

// UnbanCommand unbans one room player.
type UnbanCommand struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// RoomID identifies the packet room.
	RoomID int64
	// PlayerID identifies the target.
	PlayerID int64
}

// UnbanHandler handles room unban commands.
type UnbanHandler struct {
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings stores connection bindings.
	Bindings *binding.Registry
	// Moderation manages room moderation.
	Moderation roommoderation.Manager
}

// CommandName returns the stable command name.
func (UnbanCommand) CommandName() command.Name { return UnbanName }

// Handle commits an unban for the actor's current room.
func (handler UnbanHandler) Handle(ctx context.Context, envelope command.Envelope[UnbanCommand]) error {
	player, roomID, err := control.Actor(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	if err := control.MatchRoom(roomID, envelope.Command.RoomID); err != nil {
		return err
	}

	return handler.Moderation.Unban(ctx, roomID, player.ID(), envelope.Command.PlayerID)
}
