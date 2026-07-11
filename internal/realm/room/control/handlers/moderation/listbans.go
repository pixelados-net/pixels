package moderation

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	moderationcmd "github.com/niflaot/pixels/internal/realm/room/control/commands/moderation"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inlistbans "github.com/niflaot/pixels/networking/inbound/room/moderation/listbans"
	"go.uber.org/zap"
)

// NewBanList creates a room ban-list packet handler.
func NewBanList(handler moderationcmd.BanListHandler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inlistbans.Decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[moderationcmd.BanListCommand]{Command: moderationcmd.BanListCommand{Handler: connection, RoomID: int64(payload.RoomID)}})
	}
}

// RegisterBanList adds the room ban-list handler.
func RegisterBanList(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inlistbans.Header, handler)
}
