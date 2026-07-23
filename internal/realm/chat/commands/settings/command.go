// Package settings returns the authenticated player's Nitro user settings.
package settings

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomsession "github.com/niflaot/pixels/internal/realm/room/runtime/commands/session"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outsettings "github.com/niflaot/pixels/networking/outbound/chat/settings"
)

const (
	// Name identifies the chat user settings command.
	Name command.Name = "chat.settings"
)

// Command requests the current user settings snapshot.
type Command struct {
	// Handler stores the source connection.
	Handler netconn.Context
}

// Handler executes user settings requests.
type Handler struct {
	// Players stores live player state.
	Players *playerlive.Registry
	// Bindings resolves authenticated sessions.
	Bindings *binding.Registry
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name { return Name }

// Handle sends the current live cross-capability settings snapshot.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	player, err := roomsession.Player(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	snapshot := player.Snapshot()
	packet, err := outsettings.Encode(
		snapshot.VolumeSystem,
		snapshot.VolumeFurniture,
		snapshot.VolumeTrax,
		snapshot.OldChat,
		snapshot.BlockRoomInvites,
		snapshot.CameraFollowBlocked,
		0,
		snapshot.BubbleStyle,
	)
	if err != nil {
		return err
	}

	return envelope.Command.Handler.Send(ctx, packet)
}
