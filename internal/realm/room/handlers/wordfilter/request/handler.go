// Package request adapts the room word filter request packet.
package request

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	requestcmd "github.com/niflaot/pixels/internal/realm/room/commands/wordfilter/request"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inrequest "github.com/niflaot/pixels/networking/inbound/room/wordfilter/request"
	"go.uber.org/zap"
)

// New creates a room word filter request packet handler.
func New(handler requestcmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inrequest.Decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[requestcmd.Command]{Command: requestcmd.Command{Handler: connection, RoomID: int64(payload.RoomID)}})
	}
}

// Register adds the room word filter request handler.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inrequest.Header, handler)
}
