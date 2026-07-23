// Package typing adapts UNIT_TYPING packets to room typing commands.
package typing

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	typingcmd "github.com/niflaot/pixels/internal/realm/chat/commands/typing"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	"go.uber.org/zap"
)

// New creates a room typing packet handler for one state.
func New(handler typingcmd.Handler, active bool, decode func(codec.Packet) error, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		if err := decode(packet); err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[typingcmd.Command]{
			Command:  typingcmd.Command{Handler: connection, Active: active},
			Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)},
		})
	}
}
