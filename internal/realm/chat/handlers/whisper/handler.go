// Package whisper adapts UNIT_CHAT_WHISPER to a private chat command.
package whisper

import (
	"context"
	"strings"

	"github.com/niflaot/pixels/internal/command"
	messagecmd "github.com/niflaot/pixels/internal/realm/chat/commands/message"
	chatsend "github.com/niflaot/pixels/internal/realm/chat/send"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inwhisper "github.com/niflaot/pixels/networking/inbound/chat/whisper"
	"go.uber.org/zap"
)

// New creates a UNIT_CHAT_WHISPER packet handler.
func New(handler messagecmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inwhisper.Decode(packet)
		if err != nil {
			return err
		}
		recipient, message, _ := strings.Cut(payload.RecipientAndMessage, " ")

		return dispatcher.Dispatch(context.Background(), command.Envelope[messagecmd.Command]{
			Command:  messagecmd.Command{Handler: connection, Kind: chatsend.KindWhisper, Recipient: recipient, Message: message},
			Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)},
		})
	}
}

// Register adds UNIT_CHAT_WHISPER to a handler registry.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inwhisper.Header, handler)
}
