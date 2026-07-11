// Package typing dispatches room unit typing state.
package typing

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	chatsend "github.com/niflaot/pixels/internal/realm/chat/send"
	netconn "github.com/niflaot/pixels/networking/connection"
)

const (
	// Name identifies the room typing command.
	Name command.Name = "chat.typing"
)

// Command contains one typing-state request.
type Command struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// Active reports whether typing started.
	Active bool
}

// Handler executes typing-state commands.
type Handler struct {
	// Chat broadcasts typing state.
	Chat *chatsend.Service
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name { return Name }

// Handle executes one typing-state command.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	return handler.Chat.Typing(ctx, envelope.Command.Handler, envelope.Command.Active)
}
