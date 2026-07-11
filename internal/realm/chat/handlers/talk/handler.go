// Package talk adapts the UNIT_CHAT packet to a chat command.
package talk

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	messagecmd "github.com/niflaot/pixels/internal/realm/chat/commands/message"
	chatsend "github.com/niflaot/pixels/internal/realm/chat/send"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	intalk "github.com/niflaot/pixels/networking/inbound/chat/talk"
	"go.uber.org/zap"
)

// New creates a UNIT_CHAT packet handler.
func New(handler messagecmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := intalk.Decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[messagecmd.Command]{
			Command:  messagecmd.Command{Handler: connection, Kind: chatsend.KindTalk, Message: payload.Message},
			Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)},
		})
	}
}

// Register adds UNIT_CHAT to a handler registry.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(intalk.Header, handler)
}
