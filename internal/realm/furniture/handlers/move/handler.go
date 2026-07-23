// Package move contains the furniture move packet handler.
package move

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	movecmd "github.com/niflaot/pixels/internal/realm/furniture/commands/move"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	infloorupdate "github.com/niflaot/pixels/networking/inbound/furniture/floorupdate"
	inwallupdate "github.com/niflaot/pixels/networking/inbound/furniture/wallupdate"
	"go.uber.org/zap"
)

// New creates a furniture move packet handler.
func New(handler movecmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := infloorupdate.Decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[movecmd.Command]{
			Command: movecmd.Command{
				Handler: connection, ItemID: int64(payload.ItemID),
				X: int(payload.X), Y: int(payload.Y), Rotation: int(payload.Rotation),
			},
			Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)},
		})
	}
}

// NewWall creates a wall furniture move packet handler.
func NewWall(handler movecmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inwallupdate.Decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[movecmd.Command]{
			Command:  movecmd.Command{Handler: connection, ItemID: int64(payload.ItemID), WallPosition: payload.WallPosition},
			Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)},
		})
	}
}

// Register adds the furniture move handler to a registry.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(infloorupdate.Header, handler)
}

// RegisterWall adds the wall furniture move handler to a registry.
func RegisterWall(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inwallupdate.Header, handler)
}
