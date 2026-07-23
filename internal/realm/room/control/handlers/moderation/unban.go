package moderation

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	moderationcmd "github.com/niflaot/pixels/internal/realm/room/control/commands/moderation"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inunban "github.com/niflaot/pixels/networking/inbound/room/moderation/unban"
	"go.uber.org/zap"
)

// NewUnban creates a room unban packet handler.
func NewUnban(handler moderationcmd.UnbanHandler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inunban.Decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[moderationcmd.UnbanCommand]{Command: moderationcmd.UnbanCommand{Handler: connection, RoomID: int64(payload.RoomID), PlayerID: int64(payload.PlayerID)}})
	}
}

// RegisterUnban adds the room unban handler.
func RegisterUnban(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inunban.Header, handler)
}
