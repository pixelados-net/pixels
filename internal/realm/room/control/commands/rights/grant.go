// Package rights contains room build-right commands.
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
	// GrantName identifies the room rights grant command.
	GrantName command.Name = "room.rights.grant"
)

// GrantCommand grants room rights to one player.
type GrantCommand struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// PlayerID identifies the target.
	PlayerID int64
}

// GrantHandler handles room rights grants.
type GrantHandler struct {
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings stores connection bindings.
	Bindings *binding.Registry
	// Rights manages room rights.
	Rights roomrights.Manager
}

// CommandName returns the stable command name.
func (GrantCommand) CommandName() command.Name { return GrantName }

// Handle grants room rights.
func (handler GrantHandler) Handle(ctx context.Context, envelope command.Envelope[GrantCommand]) error {
	player, roomID, err := control.Actor(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}

	return handler.Rights.GrantRights(ctx, roomID, player.ID(), envelope.Command.PlayerID)
}
