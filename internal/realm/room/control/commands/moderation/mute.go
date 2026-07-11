package moderation

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/control/commands/resolve"
	roommoderation "github.com/niflaot/pixels/internal/realm/room/control/moderation"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
)

const (
	// MuteName identifies the room mute command.
	MuteName command.Name = "room.moderation.mute"
)

// MuteCommand mutes one room occupant.
type MuteCommand struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// RoomID identifies the packet room.
	RoomID int64
	// PlayerID identifies the target.
	PlayerID int64
	// Minutes stores the requested mute duration.
	Minutes int32
}

// MuteHandler handles room mute commands.
type MuteHandler struct {
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings stores connection bindings.
	Bindings *binding.Registry
	// Moderation manages room moderation.
	Moderation roommoderation.Manager
	// Runtime stores active room presence.
	Runtime *roomlive.Registry
}

// CommandName returns the stable command name.
func (MuteCommand) CommandName() command.Name { return MuteName }

// Handle commits a mute for the actor's current room.
func (handler MuteHandler) Handle(ctx context.Context, envelope command.Envelope[MuteCommand]) error {
	player, roomID, err := control.Actor(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	if err := control.MatchRoom(roomID, envelope.Command.RoomID); err != nil {
		return err
	}
	if err := control.TargetInRoom(handler.Runtime, roomID, envelope.Command.PlayerID); err != nil {
		return err
	}

	return handler.Moderation.Mute(ctx, roomID, player.ID(), envelope.Command.PlayerID, envelope.Command.Minutes)
}
