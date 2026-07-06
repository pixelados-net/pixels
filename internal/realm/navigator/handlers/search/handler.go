// Package search contains the navigator search packet handler.
package search

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	searchcmd "github.com/niflaot/pixels/internal/realm/navigator/commands/search"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	insearch "github.com/niflaot/pixels/networking/inbound/navigator/search"
	"go.uber.org/zap"
)

// New creates a navigator search packet handler.
func New(handler searchcmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher[searchcmd.Command](handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := insearch.Decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[searchcmd.Command]{
			Command: searchcmd.Command{Handler: connection, Code: payload.Code, Data: payload.Data},
			Metadata: command.Metadata{
				ConnectionID: string(connection.ConnectionID),
			},
		})
	}
}

// Register adds the navigator search handler to a registry.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(insearch.Header, handler)
}
