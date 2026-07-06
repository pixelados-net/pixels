// Package enter contains the room enter packet handler.
package enter

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	entercmd "github.com/niflaot/pixels/internal/realm/room/commands/enter"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inenter "github.com/niflaot/pixels/networking/inbound/room/enter"
	"go.uber.org/zap"
)

// New creates a room enter packet handler.
func New(handler entercmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inenter.Decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[entercmd.Command]{
			Command:  entercmd.Command{Handler: connection, RoomID: int64(payload.FlatID), Password: payload.Password},
			Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)},
		})
	}
}

// Register adds the room enter handler to a registry.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inenter.Header, handler)
}
