// Package grant grants room build rights.
package grant

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/room/commands/control"
	roomrights "github.com/niflaot/pixels/internal/realm/room/rights"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
)

const (
	// Name identifies the room rights grant command.
	Name command.Name = "room.rights.grant"
)

// Command grants room rights to one player.
type Command struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// PlayerID identifies the target.
	PlayerID int64
}

// Handler handles room rights grants.
type Handler struct {
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings stores connection bindings.
	Bindings *binding.Registry
	// Rights manages room rights.
	Rights roomrights.Manager
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name { return Name }

// Handle grants room rights.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	player, roomID, err := control.Actor(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}

	return handler.Rights.GrantRights(ctx, roomID, player.ID(), envelope.Command.PlayerID)
}
