// Package tags contains room tag packet handlers.
package tags

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	tagscmd "github.com/niflaot/pixels/internal/realm/room/commands/tags"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	intags "github.com/niflaot/pixels/networking/inbound/room/tags"
	"go.uber.org/zap"
)

// New creates a room tags packet handler.
func New(handler tagscmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		if _, err := intags.Decode(packet); err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[tagscmd.Command]{
			Command:  tagscmd.Command{Handler: connection},
			Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)},
		})
	}
}

// Register adds room tag handlers to a registry.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(intags.SessionHeader, handler)
	_ = registry.Register(intags.PopularHeader, handler)
}
