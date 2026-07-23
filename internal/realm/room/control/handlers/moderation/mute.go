package moderation

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	moderationcmd "github.com/niflaot/pixels/internal/realm/room/control/commands/moderation"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inmute "github.com/niflaot/pixels/networking/inbound/room/moderation/mute"
	"go.uber.org/zap"
)

// NewMute creates a room mute packet handler.
func NewMute(muteHandler moderationcmd.MuteHandler, unmuteHandler moderationcmd.UnmuteHandler, log *zap.Logger) netconn.Handler {
	muteDispatcher, _ := command.NewDispatcher(muteHandler)
	muteDispatcher.WithLogger(log)
	unmuteDispatcher, _ := command.NewDispatcher(unmuteHandler)
	unmuteDispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inmute.Decode(packet)
		if err != nil {
			return err
		}
		if payload.Minutes == 0 {
			return unmuteDispatcher.Dispatch(context.Background(), command.Envelope[moderationcmd.UnmuteCommand]{Command: moderationcmd.UnmuteCommand{Handler: connection, RoomID: int64(payload.RoomID), PlayerID: int64(payload.PlayerID)}})
		}

		return muteDispatcher.Dispatch(context.Background(), command.Envelope[moderationcmd.MuteCommand]{Command: moderationcmd.MuteCommand{Handler: connection, RoomID: int64(payload.RoomID), PlayerID: int64(payload.PlayerID), Minutes: payload.Minutes}})
	}
}

// RegisterMute adds the room mute handler.
func RegisterMute(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inmute.Header, handler)
}
