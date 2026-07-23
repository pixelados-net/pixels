// Package create contains the navigator room creation packet handler.
package create

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	increate "github.com/niflaot/pixels/networking/inbound/navigator/create/room"
	"go.uber.org/zap"
)

// New creates a navigator room creation packet handler.
func NewPacketHandler(handler Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := increate.Decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[Command]{
			Command: createCommand(connection, payload),
			Metadata: command.Metadata{
				ConnectionID: string(connection.ConnectionID),
			},
		})
	}
}

// Register adds the navigator room creation handler to a registry.
func RegisterPacketHandler(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(increate.Header, handler)
}

// createCommand maps packet payload to a command.
func createCommand(connection netconn.Context, payload increate.Payload) Command {
	return Command{
		Handler:         connection,
		RoomName:        payload.RoomName,
		RoomDescription: payload.RoomDescription,
		ModelName:       payload.ModelName,
		CategoryID:      payload.CategoryID,
		MaxVisitors:     payload.MaxVisitors,
		TradeType:       payload.TradeType,
	}
}
