// Package revoke revokes room build rights.
package revoke

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
	// Name identifies the room rights revoke command.
	Name command.Name = "room.rights.revoke"
)

// Command revokes room rights from players.
type Command struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// PlayerIDs identifies the targets.
	PlayerIDs []int64
}

// Handler handles room rights revocations.
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

// Handle revokes each requested rights holder.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	player, roomID, err := control.Actor(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	for _, playerID := range envelope.Command.PlayerIDs {
		if err := handler.Rights.RevokeRights(ctx, roomID, player.ID(), playerID); err != nil {
			return err
		}
	}

	return nil
}
