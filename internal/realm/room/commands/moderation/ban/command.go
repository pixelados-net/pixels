// Package ban bans a player from the actor's current room.
package ban

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/commands/control"
	roomlive "github.com/niflaot/pixels/internal/realm/room/live"
	roommoderation "github.com/niflaot/pixels/internal/realm/room/moderation"
	moderationmodel "github.com/niflaot/pixels/internal/realm/room/moderation/model"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
)

const (
	// Name identifies the room ban command.
	Name command.Name = "room.moderation.ban"
)

// Command bans one room player.
type Command struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// RoomID identifies the packet room.
	RoomID int64
	// PlayerID identifies the target.
	PlayerID int64
	// Duration stores the Nitro ban duration.
	Duration moderationmodel.BanDuration
}

// Handler handles room ban commands.
type Handler struct {
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
func (Command) CommandName() command.Name { return Name }

// Handle commits a ban for the actor's current room.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
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

	return handler.Moderation.Ban(ctx, roomID, player.ID(), envelope.Command.PlayerID, envelope.Command.Duration)
}
