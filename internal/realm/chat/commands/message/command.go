// Package message dispatches decoded room chat messages.
package message

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	chatsend "github.com/niflaot/pixels/internal/realm/chat/send"
	netconn "github.com/niflaot/pixels/networking/connection"
)

const (
	// Name identifies the room chat message command.
	Name command.Name = "chat.message"
)

// Command contains one decoded chat request.
type Command struct {
	// Handler stores the source connection.
	Handler netconn.Context
	// Kind stores talk, shout, or whisper semantics.
	Kind chatsend.Kind
	// Message stores submitted text.
	Message string
	// Recipient stores a whisper recipient username.
	Recipient string
}

// Handler executes room chat message commands.
type Handler struct {
	// Chat validates and delivers messages.
	Chat *chatsend.Service
}

// CommandName returns the stable command name.
func (Command) CommandName() command.Name { return Name }

// Handle executes one room chat message command.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	value := envelope.Command
	return handler.Chat.Handle(ctx, chatsend.Request{Handler: value.Handler, Kind: value.Kind, Message: value.Message, Recipient: value.Recipient})
}
