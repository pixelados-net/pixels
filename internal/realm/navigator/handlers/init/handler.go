// Package init contains the navigator init packet handler.
package init

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	initcmd "github.com/niflaot/pixels/internal/realm/navigator/commands/init"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	innav "github.com/niflaot/pixels/networking/inbound/navigator/init"
	"go.uber.org/zap"
)

// New creates a navigator init packet handler.
func New(handler initcmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher[initcmd.Command](handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		if _, err := innav.Decode(packet); err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[initcmd.Command]{
			Command:  initcmd.Command{Handler: connection},
			Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)},
		})
	}
}

// Register adds the navigator init handler to a registry.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(innav.Header, handler)
}
