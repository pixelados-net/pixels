// Package modify adapts the room word filter modify packet.
package modify

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	modifycmd "github.com/niflaot/pixels/internal/realm/room/commands/wordfilter/modify"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inmodify "github.com/niflaot/pixels/networking/inbound/room/wordfilter/modify"
	"go.uber.org/zap"
)

// New creates a room word filter modify packet handler.
func New(handler modifycmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inmodify.Decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[modifycmd.Command]{Command: modifycmd.Command{Handler: connection, RoomID: int64(payload.RoomID), Add: payload.Add, Word: payload.Word}})
	}
}

// Register adds the room word filter modify handler.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inmodify.Header, handler)
}
