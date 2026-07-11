// Package style dispatches validated bubble style selections.
package style

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	"github.com/niflaot/pixels/internal/realm/chat/bubble"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomsession "github.com/niflaot/pixels/internal/realm/room/commands/session"
	"github.com/niflaot/pixels/internal/realm/session/binding"
	netconn "github.com/niflaot/pixels/networking/connection"
)

const (
	// Name identifies the chat bubble style command.
	Name command.Name = "chat.style"
)

// Command contains one bubble style selection.
type Command struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// StyleID identifies the requested Nitro bubble.
	StyleID int32
}

// Handler executes bubble style selection commands.
type Handler struct {
	// Players stores live player state.
	Players *playerlive.Registry
	// Bindings resolves source sessions.
	Bindings *binding.Registry
	// Bubbles validates and persists styles.
	Bubbles *bubble.Service
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name { return Name }

// Handle executes one bubble style selection.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	player, err := roomsession.Player(envelope.Command.Handler, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}

	return handler.Bubbles.Select(ctx, player.ID(), envelope.Command.StyleID)
}
