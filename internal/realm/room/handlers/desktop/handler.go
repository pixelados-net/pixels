// Package desktop contains the desktop view room leave handler.
package desktop

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	leavecmd "github.com/niflaot/pixels/internal/realm/room/commands/leave"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	indesktop "github.com/niflaot/pixels/networking/inbound/session/desktop"
	"go.uber.org/zap"
)

// New creates a desktop view packet handler.
func New(handler leavecmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		if _, err := indesktop.Decode(packet); err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[leavecmd.Command]{
			Command:  leavecmd.Command{Handler: connection},
			Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)},
		})
	}
}

// Register adds the desktop view handler to a registry.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(indesktop.Header, handler)
}
