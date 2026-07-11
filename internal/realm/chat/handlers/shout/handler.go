// Package shout adapts the UNIT_CHAT_SHOUT packet to a chat command.
package shout

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	messagecmd "github.com/niflaot/pixels/internal/realm/chat/commands/message"
	chatsend "github.com/niflaot/pixels/internal/realm/chat/send"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inshout "github.com/niflaot/pixels/networking/inbound/chat/shout"
	"go.uber.org/zap"
)

// New creates a UNIT_CHAT_SHOUT packet handler.
func New(handler messagecmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inshout.Decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[messagecmd.Command]{
			Command:  messagecmd.Command{Handler: connection, Kind: chatsend.KindShout, Message: payload.Message},
			Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)},
		})
	}
}

// Register adds UNIT_CHAT_SHOUT to a handler registry.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inshout.Header, handler)
}
