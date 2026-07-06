// Package cancreate contains the navigator can-create packet handler.
package cancreate

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	cancreatecmd "github.com/niflaot/pixels/internal/realm/navigator/commands/cancreate"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	incancreate "github.com/niflaot/pixels/networking/inbound/navigator/cancreate"
	"go.uber.org/zap"
)

// New creates a navigator can-create packet handler.
func New(handler cancreatecmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher[cancreatecmd.Command](handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		if _, err := incancreate.Decode(packet); err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[cancreatecmd.Command]{
			Command:  cancreatecmd.Command{Handler: connection},
			Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)},
		})
	}
}

// Register adds the navigator can-create handler to a registry.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(incancreate.Header, handler)
}
