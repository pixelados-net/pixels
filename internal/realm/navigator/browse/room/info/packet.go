// Package info contains the navigator room info packet handler.
package info

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	ininfo "github.com/niflaot/pixels/networking/inbound/navigator/browse/roominfo"
	"go.uber.org/zap"
)

// New creates a navigator room info packet handler.
func NewPacketHandler(handler Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := ininfo.Decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[Command]{
			Command: roomInfoCommand(connection, payload),
			Metadata: command.Metadata{
				ConnectionID: string(connection.ConnectionID),
			},
		})
	}
}

// Register adds the navigator room info handler to a registry.
func RegisterPacketHandler(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(ininfo.Header, handler)
}

// roomInfoCommand maps packet payload to command input.
func roomInfoCommand(connection netconn.Context, payload ininfo.Payload) Command {
	return Command{
		Handler:     connection,
		RoomID:      int64(payload.RoomID),
		EnterRoom:   payload.EnterRoom > 0,
		ForwardRoom: payload.ForwardRoom > 0,
	}
}
