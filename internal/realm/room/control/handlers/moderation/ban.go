// Package moderation adapts room moderation packets into commands.
package moderation

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	moderationcmd "github.com/niflaot/pixels/internal/realm/room/control/commands/moderation"
	moderationmodel "github.com/niflaot/pixels/internal/realm/room/control/moderation/model"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inban "github.com/niflaot/pixels/networking/inbound/room/moderation/ban"
	"go.uber.org/zap"
)

// NewBan creates a room ban packet handler.
func NewBan(handler moderationcmd.BanHandler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inban.Decode(packet)
		if err != nil {
			return err
		}
		roomCommand := moderationcmd.BanCommand{Handler: connection, RoomID: int64(payload.RoomID), PlayerID: int64(payload.PlayerID), Duration: moderationmodel.BanDuration(payload.Duration)}

		return dispatcher.Dispatch(context.Background(), command.Envelope[moderationcmd.BanCommand]{Command: roomCommand})
	}
}

// RegisterBan adds the room ban handler.
func RegisterBan(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inban.Header, handler)
}
