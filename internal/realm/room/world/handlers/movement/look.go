package movement

import (
	"context"
	"github.com/niflaot/pixels/internal/command"
	lookcmd "github.com/niflaot/pixels/internal/realm/room/world/commands/look"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inlook "github.com/niflaot/pixels/networking/inbound/room/entities/look"
	"go.uber.org/zap"
)

// NewLook creates a room look packet handler.
func NewLook(handler lookcmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inlook.Decode(packet)
		if err != nil {
			return err
		}

		err = dispatcher.Dispatch(context.Background(), command.Envelope[lookcmd.Command]{
			Command:  lookcmd.Command{Handler: connection, X: int(payload.X), Y: int(payload.Y)},
			Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)},
		})
		if isStalePresenceError(err) {
			if log != nil {
				log.Debug("room look ignored after presence ended", zap.Error(err), zap.String("connection_id", string(connection.ConnectionID)))
			}
			return nil
		}

		return err
	}
}

// RegisterLook adds the room look handler to a registry.
func RegisterLook(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inlook.Header, handler)
}
