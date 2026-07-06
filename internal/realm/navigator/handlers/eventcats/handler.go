// Package eventcats contains the navigator event categories packet handler.
package eventcats

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	eventcatscmd "github.com/niflaot/pixels/internal/realm/navigator/commands/eventcats"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	ineventcats "github.com/niflaot/pixels/networking/inbound/navigator/eventcats"
	"go.uber.org/zap"
)

// New creates a navigator event categories packet handler.
func New(handler eventcatscmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		if _, err := ineventcats.Decode(packet); err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[eventcatscmd.Command]{
			Command:  eventcatscmd.Command{Handler: connection},
			Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)},
		})
	}
}

// Register adds the navigator event categories handler to a registry.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(ineventcats.Header, handler)
}
