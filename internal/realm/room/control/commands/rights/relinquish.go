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
	// RelinquishName identifies the room rights relinquish command.
	RelinquishName command.Name = "room.rights.relinquish"
)

// RelinquishCommand relinquishes the actor's room rights.
type RelinquishCommand struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// RoomID identifies the packet room.
	RoomID int64
}

// RelinquishHandler handles rights relinquishment.
type RelinquishHandler struct {
	// Players stores live players.
	Players *playerlive.Registry
	// Bindings stores connection bindings.
	Bindings *binding.Registry
	// Rights manages room rights.
	Rights roomrights.Manager
}

// CommandName returns the stable command name.
func (RelinquishCommand) CommandName() command.Name { return RelinquishName }

// Handle relinquishes rights in the current room.
func (handler RelinquishHandler) Handle(ctx context.Context, envelope command.Envelope[RelinquishCommand]) error {
	player, roomID, err := control.Actor(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	if err := control.MatchRoom(roomID, envelope.Command.RoomID); err != nil {
		return err
	}

	return handler.Rights.RelinquishRights(ctx, roomID, player.ID())
}
