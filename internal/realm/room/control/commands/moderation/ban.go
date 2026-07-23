// Package moderation contains room moderation commands.
package moderation

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/control/commands/resolve"
	roommoderation "github.com/niflaot/pixels/internal/realm/room/control/moderation"
	moderationmodel "github.com/niflaot/pixels/internal/realm/room/control/moderation/model"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
)

const (
	// BanName identifies the room ban command.
	BanName command.Name = "room.moderation.ban"
)

// BanCommand bans one room player.
type BanCommand struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// RoomID identifies the packet room.
	RoomID int64
	// PlayerID identifies the target.
	PlayerID int64
	// Duration stores the Nitro ban duration.
	Duration moderationmodel.BanDuration
}

// BanHandler handles room ban commands.
type BanHandler struct {
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
func (BanCommand) CommandName() command.Name { return BanName }

// Handle commits a ban for the actor's current room.
func (handler BanHandler) Handle(ctx context.Context, envelope command.Envelope[BanCommand]) error {
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
