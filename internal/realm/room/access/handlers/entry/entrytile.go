package entry

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	entrytilecmd "github.com/niflaot/pixels/internal/realm/room/access/commands/entrytile"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inentrytile "github.com/niflaot/pixels/networking/inbound/room/entrytile"
	"go.uber.org/zap"
)

// NewEntryTile creates a room entry tile packet handler.
func NewEntryTile(handler entrytilecmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		if _, err := inentrytile.Decode(packet); err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[entrytilecmd.Command]{
			Command:  entrytilecmd.Command{Handler: connection},
			Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)},
		})
	}
}

// RegisterEntryTile adds the room entry tile handler to a registry.
func RegisterEntryTile(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inentrytile.Header, handler)
}
