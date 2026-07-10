// Package unban adapts the room unban packet.
package unban

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	unbancmd "github.com/niflaot/pixels/internal/realm/room/commands/moderation/unban"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inunban "github.com/niflaot/pixels/networking/inbound/room/moderation/unban"
	"go.uber.org/zap"
)

// New creates a room unban packet handler.
func New(handler unbancmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inunban.Decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[unbancmd.Command]{Command: unbancmd.Command{Handler: connection, RoomID: int64(payload.RoomID), PlayerID: int64(payload.PlayerID)}})
	}
}

// Register adds the room unban handler.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inunban.Header, handler)
}
