// Package categorycounts contains the navigator category counts packet handler.
package categorycounts

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	incounts "github.com/niflaot/pixels/networking/inbound/navigator/browse/categorycounts"
	"go.uber.org/zap"
)

// New creates a navigator category counts packet handler.
func NewPacketHandler(handler Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		if _, err := incounts.Decode(packet); err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[Command]{
			Command:  Command{Handler: connection},
			Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)},
		})
	}
}

// Register adds the navigator category counts handler to a registry.
func RegisterPacketHandler(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(incounts.Header, handler)
}
