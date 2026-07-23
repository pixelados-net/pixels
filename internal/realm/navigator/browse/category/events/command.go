// Package eventcats sends available event categories to navigator clients.
package eventcats

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	navsession "github.com/niflaot/pixels/internal/realm/navigator/session"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outeventcats "github.com/niflaot/pixels/networking/outbound/navigator/browse/eventcategories"
)

const (
	// Name identifies the navigator event categories command.
	Name command.Name = "navigator.event_cats"
)

// Command sends navigator event categories.
type Command struct {
	// Handler stores the source connection handler.
	Handler netconn.Context
}

// Handler handles event category requests.
type Handler struct {
	// Players stores live player state.
	Players *playerlive.Registry
	// Bindings stores player connection bindings.
	Bindings *binding.Registry
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name {
	return Name
}

// Handle handles an event category command.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	if _, _, err := navsession.Player(envelope.Command.Handler, handler.Bindings, handler.Players); err != nil {
		return err
	}

	packet, err := outeventcats.Encode(nil)
	if err != nil {
		return err
	}

	return envelope.Command.Handler.Send(ctx, packet)
}
