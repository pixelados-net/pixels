package moderation

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	moderationcmd "github.com/niflaot/pixels/internal/realm/room/control/commands/moderation"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inkick "github.com/niflaot/pixels/networking/inbound/room/moderation/kick"
	"go.uber.org/zap"
)

// NewKick creates a room kick packet handler.
func NewKick(handler moderationcmd.KickHandler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inkick.Decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[moderationcmd.KickCommand]{Command: moderationcmd.KickCommand{Handler: connection, PlayerID: int64(payload.PlayerID)}})
	}
}

// RegisterKick adds the room kick handler.
func RegisterKick(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inkick.Header, handler)
}
