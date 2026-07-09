// Package request contains the currency wallet request packet handler.
package request

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	requestcmd "github.com/niflaot/pixels/internal/realm/inventory/currency/commands/request"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inrequest "github.com/niflaot/pixels/networking/inbound/user/currency/request"
	"go.uber.org/zap"
)

// New creates a currency wallet request packet handler.
func New(handler *requestcmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		if _, err := inrequest.Decode(packet); err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[requestcmd.Command]{
			Command:  requestcmd.Command{Connection: connection},
			Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)},
		})
	}
}

// Register adds the currency wallet request handler to a registry.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inrequest.Header, handler)
}
