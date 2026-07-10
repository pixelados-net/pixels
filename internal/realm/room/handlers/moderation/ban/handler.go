// Package ban adapts the room ban packet.
package ban

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	bancmd "github.com/niflaot/pixels/internal/realm/room/commands/moderation/ban"
	moderationmodel "github.com/niflaot/pixels/internal/realm/room/moderation/model"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inban "github.com/niflaot/pixels/networking/inbound/room/moderation/ban"
	"go.uber.org/zap"
)

// New creates a room ban packet handler.
func New(handler bancmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inban.Decode(packet)
		if err != nil {
			return err
		}
		roomCommand := bancmd.Command{Handler: connection, RoomID: int64(payload.RoomID), PlayerID: int64(payload.PlayerID), Duration: moderationmodel.BanDuration(payload.Duration)}

		return dispatcher.Dispatch(context.Background(), command.Envelope[bancmd.Command]{Command: roomCommand})
	}
}

// Register adds the room ban handler.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inban.Header, handler)
}
