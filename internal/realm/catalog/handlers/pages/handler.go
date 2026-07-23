// Package pages adapts GET_CATALOG_INDEX packets to catalog commands.
package pages

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	pagescmd "github.com/niflaot/pixels/internal/realm/catalog/commands/pages"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inpages "github.com/niflaot/pixels/networking/inbound/catalog/mode/request"
	"go.uber.org/zap"
)

// New creates a catalog pages packet handler.
func New(handler pagescmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inpages.Decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[pagescmd.Command]{
			Command:  pagescmd.Command{Connection: connection, Mode: payload.Mode},
			Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)},
		})
	}
}

// Register adds the catalog pages handler to a registry.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inpages.Header, handler)
}
