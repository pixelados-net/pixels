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
	// KickName identifies the room kick command.
	KickName command.Name = "room.moderation.kick"
)

// KickCommand kicks one room occupant.
type KickCommand struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// PlayerID identifies the target.
	PlayerID int64
}

// KickHandler handles room kick commands.
type KickHandler struct {
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
func (KickCommand) CommandName() command.Name { return KickName }

// Handle commits a kick for the actor's current room.
func (handler KickHandler) Handle(ctx context.Context, envelope command.Envelope[KickCommand]) error {
	player, roomID, err := control.Actor(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	if err := control.TargetInRoom(handler.Runtime, roomID, envelope.Command.PlayerID); err != nil {
		return err
	}

	return handler.Moderation.Kick(ctx, roomID, player.ID(), envelope.Command.PlayerID)
}
