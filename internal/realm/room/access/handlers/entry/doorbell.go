package entry

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	respondcmd "github.com/niflaot/pixels/internal/realm/room/access/commands/doorbell/respond"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inrespond "github.com/niflaot/pixels/networking/inbound/room/doorbell/respond"
	"go.uber.org/zap"
)

// NewDoorbell creates a room doorbell packet handler.
func NewDoorbell(handler respondcmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inrespond.Decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[respondcmd.Command]{
			Command:  respondcmd.Command{Handler: connection, Username: payload.Username, Accepted: payload.Accepted},
			Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)},
		})
	}
}

// RegisterDoorbell adds the room doorbell handler to a registry.
func RegisterDoorbell(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inrespond.Header, handler)
}
