// Package forward contains the navigator forward packet handler.
package forward

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	forwardcmd "github.com/niflaot/pixels/internal/realm/navigator/commands/forward"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inforward "github.com/niflaot/pixels/networking/inbound/navigator/forward"
	"go.uber.org/zap"
)

// New creates a navigator forward packet handler.
func New(handler forwardcmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		if _, err := inforward.Decode(packet); err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[forwardcmd.Command]{
			Command:  forwardcmd.Command{Handler: connection},
			Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)},
		})
	}
}

// Register adds the navigator forward handler to a registry.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inforward.Header, handler)
}
