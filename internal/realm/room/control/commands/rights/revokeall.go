package rights

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/control/commands/resolve"
	roomrights "github.com/niflaot/pixels/internal/realm/room/control/rights"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
)

const (
	// RevokeAllName identifies the room rights revoke-all command.
	RevokeAllName command.Name = "room.rights.revoke_all"
)

// RevokeAllCommand revokes every room rights holder.
type RevokeAllCommand struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// RoomID identifies the packet room.
	RoomID int64
}

// RevokeAllHandler handles revoke-all requests.
type RevokeAllHandler struct {
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings stores connection bindings.
	Bindings *binding.Registry
	// Rights manages room rights.
	Rights roomrights.Manager
}

// CommandName returns the stable command name.
func (RevokeAllCommand) CommandName() command.Name { return RevokeAllName }

// Handle revokes all rights in the current room.
func (handler RevokeAllHandler) Handle(ctx context.Context, envelope command.Envelope[RevokeAllCommand]) error {
	player, roomID, err := control.Actor(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	if err := control.MatchRoom(roomID, envelope.Command.RoomID); err != nil {
		return err
	}
	_, err = handler.Rights.RevokeAllRights(ctx, roomID, player.ID())

	return err
}
