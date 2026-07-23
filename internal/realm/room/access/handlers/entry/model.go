package entry

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	modelcmd "github.com/niflaot/pixels/internal/realm/room/access/commands/model"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inmodel "github.com/niflaot/pixels/networking/inbound/room/model"
	"go.uber.org/zap"
)

// NewModel creates a room model packet handler.
func NewModel(handler modelcmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		if _, err := inmodel.Decode(packet); err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[modelcmd.Command]{
			Command:  modelcmd.Command{Handler: connection},
			Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)},
		})
	}
}

// RegisterModel adds the room model handler to a registry.
func RegisterModel(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inmodel.Header, handler)
}
