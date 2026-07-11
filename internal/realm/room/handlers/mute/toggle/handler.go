// Package toggle adapts the room mute-all toggle packet.
package toggle

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	togglecmd "github.com/niflaot/pixels/internal/realm/room/commands/mute/toggle"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	intoggle "github.com/niflaot/pixels/networking/inbound/room/mute/toggle"
	"go.uber.org/zap"
)

// New creates a room mute-all toggle packet handler.
func New(handler togglecmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		if err := intoggle.Decode(packet); err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[togglecmd.Command]{Command: togglecmd.Command{Handler: connection}})
	}
}

// Register adds the room mute-all toggle handler.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(intoggle.Header, handler)
}
