// Package categorycounts sends navigator category occupancy counts.
package categorycounts

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	navsession "github.com/niflaot/pixels/internal/realm/navigator/session"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
	outcounts "github.com/niflaot/pixels/networking/outbound/navigator/browse/categorycounts"
)

const (
	// Name identifies the navigator category counts command.
	Name command.Name = "navigator.category_counts"
)

// Reader reads current category count entries.
type Reader interface {
	// Snapshot returns current category count entries.
	Snapshot() []outcounts.Entry
}

// Command sends current navigator category counts.
type Command struct {
	// Handler stores the source connection handler.
	Handler netconn.Context
}

// Handler handles category count requests.
type Handler struct {
	// Players stores live player state.
	Players *playerlive.Registry
	// Bindings stores player connection bindings.
	Bindings *binding.Registry
	// Counts reads current category occupancy counts.
	Counts Reader
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name {
	return Name
}

// Handle handles a category counts command.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	if _, _, err := navsession.Player(envelope.Command.Handler, handler.Bindings, handler.Players); err != nil {
		return err
	}

	packet, err := outcounts.Encode(handler.entries())
	if err != nil {
		return err
	}

	return envelope.Command.Handler.Send(ctx, packet)
}

// entries returns current counts.
func (handler Handler) entries() []outcounts.Entry {
	if handler.Counts == nil {
		return nil
	}

	return handler.Counts.Snapshot()
}
