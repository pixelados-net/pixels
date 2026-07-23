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
	// RevokeName identifies the room rights revoke command.
	RevokeName command.Name = "room.rights.revoke"
)

// RevokeCommand revokes room rights from players.
type RevokeCommand struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// PlayerIDs identifies the targets.
	PlayerIDs []int64
}

// RevokeHandler handles room rights revocations.
type RevokeHandler struct {
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings stores connection bindings.
	Bindings *binding.Registry
	// Rights manages room rights.
	Rights roomrights.Manager
}

// CommandName returns the stable command name.
func (RevokeCommand) CommandName() command.Name { return RevokeName }

// Handle revokes each requested rights holder.
func (handler RevokeHandler) Handle(ctx context.Context, envelope command.Envelope[RevokeCommand]) error {
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
