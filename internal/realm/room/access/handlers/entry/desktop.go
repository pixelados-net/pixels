package entry

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	leavecmd "github.com/niflaot/pixels/internal/realm/room/access/commands/leave"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	indesktop "github.com/niflaot/pixels/networking/inbound/session/desktop"
	"go.uber.org/zap"
)

// NewDesktop creates a desktop view packet handler.
func NewDesktop(handler leavecmd.Handler, log *zap.Logger) netconn.Handler {
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

// RegisterDesktop adds the desktop view handler to a registry.
func RegisterDesktop(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(indesktop.Header, handler)
}
