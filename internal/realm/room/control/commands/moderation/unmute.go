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
	// UnmuteName identifies the room unmute command.
	UnmuteName command.Name = "room.moderation.unmute"
)

// UnmuteCommand unmutes one room occupant.
type UnmuteCommand struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// RoomID identifies the packet room.
	RoomID int64
	// PlayerID identifies the target.
	PlayerID int64
}

// UnmuteHandler handles room unmute commands.
type UnmuteHandler struct {
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings stores connection bindings.
	Bindings *binding.Registry
	// Moderation manages room moderation.
	Moderation roommoderation.Manager
}

// CommandName returns the stable command name.
func (UnmuteCommand) CommandName() command.Name { return UnmuteName }

// Handle commits an unmute for the actor's current room.
func (handler UnmuteHandler) Handle(ctx context.Context, envelope command.Envelope[UnmuteCommand]) error {
	player, roomID, err := control.Actor(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	if err := control.MatchRoom(roomID, envelope.Command.RoomID); err != nil {
		return err
	}

	return handler.Moderation.Unmute(ctx, roomID, player.ID(), envelope.Command.PlayerID)
}
