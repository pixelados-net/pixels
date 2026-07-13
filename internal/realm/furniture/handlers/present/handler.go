// Package present contains present packet handlers.
package present

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	presentcmd "github.com/niflaot/pixels/internal/realm/furniture/commands/present"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inopen "github.com/niflaot/pixels/networking/inbound/furniture/present/open"
	"go.uber.org/zap"
)

// New creates a present open packet handler.
func New(handler presentcmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inopen.Decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[presentcmd.Command]{
			Command:  presentcmd.Command{Handler: connection, ItemID: int64(payload.ItemID)},
			Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)},
		})
	}
}

// Register adds the present open handler to a registry.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inopen.Header, handler)
}
