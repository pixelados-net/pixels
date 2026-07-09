// Package inventory contains the furniture inventory list packet handler.
package inventory

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	inventorycmd "github.com/niflaot/pixels/internal/realm/furniture/commands/inventory"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	ininventory "github.com/niflaot/pixels/networking/inbound/inventory/furniture"
	"go.uber.org/zap"
)

// New creates a furniture inventory list packet handler.
func New(handler inventorycmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		if _, err := ininventory.Decode(packet); err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[inventorycmd.Command]{
			Command:  inventorycmd.Command{Handler: connection},
			Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)},
		})
	}
}

// Register adds the furniture inventory list handler to a registry.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(ininventory.Header, handler)
}
