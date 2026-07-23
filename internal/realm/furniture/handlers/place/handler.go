// Package place contains the furniture place packet handler.
package place

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	placecmd "github.com/niflaot/pixels/internal/realm/furniture/commands/place"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inplace "github.com/niflaot/pixels/networking/inbound/furniture/place"
	"go.uber.org/zap"
)

// New creates a furniture place packet handler.
func New(handler placecmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inplace.Decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[placecmd.Command]{
			Command: placecmd.Command{
				Handler: connection, ItemID: int64(payload.ItemID),
				X: int(payload.X), Y: int(payload.Y), Rotation: int(payload.Rotation), WallPosition: payload.WallPosition,
			},
			Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)},
		})
	}
}

// Register adds the furniture place handler to a registry.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inplace.Header, handler)
}
