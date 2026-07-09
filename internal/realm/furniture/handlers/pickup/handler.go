// Package pickup contains the furniture pickup packet handler.
package pickup

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	pickupcmd "github.com/niflaot/pixels/internal/realm/furniture/commands/pickup"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inpickup "github.com/niflaot/pixels/networking/inbound/furniture/pickup"
	"go.uber.org/zap"
)

// New creates a furniture pickup packet handler.
func New(handler pickupcmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inpickup.Decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[pickupcmd.Command]{
			Command:  pickupcmd.Command{Handler: connection, ItemID: int64(payload.ItemID)},
			Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)},
		})
	}
}

// Register adds the furniture pickup handler to a registry.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inpickup.Header, handler)
}
