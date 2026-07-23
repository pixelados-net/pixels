// Package forward sends clients to an available room.
package forward

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	navsession "github.com/niflaot/pixels/internal/realm/navigator/session"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outnosuch "github.com/niflaot/pixels/networking/outbound/navigator/browse/nosuchflat"
	outforward "github.com/niflaot/pixels/networking/outbound/room/forward"
)

const (
	// Name identifies the navigator forward command.
	Name command.Name = "navigator.forward"
)

// Command forwards a player to an available room.
type Command struct {
	// Handler stores the source connection handler.
	Handler netconn.Context
}

// Handler handles navigator forward commands.
type Handler struct {
	// Players stores live player state.
	Players *playerlive.Registry
	// Bindings stores player connection bindings.
	Bindings *binding.Registry
	// Rooms reads room persistence.
	Rooms roomservice.Manager
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name {
	return Name
}

// Handle handles a navigator forward command.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	if _, _, err := navsession.Player(envelope.Command.Handler, handler.Bindings, handler.Players); err != nil {
		return err
	}

	rooms, err := handler.Rooms.ListPopular(ctx, 1)
	if err != nil {
		return err
	}
	if len(rooms) == 0 {
		return sendMissing(ctx, envelope.Command.Handler)
	}

	packet, err := outforward.Encode(int32(rooms[0].ID))
	if err != nil {
		return err
	}

	return envelope.Command.Handler.Send(ctx, packet)
}

// sendMissing sends a no-room response.
func sendMissing(ctx context.Context, connection netconn.Context) error {
	packet, err := outnosuch.Encode(0)
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
}
