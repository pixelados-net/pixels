// Package mute adapts room mute and unmute packets.
package mute

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	mutecmd "github.com/niflaot/pixels/internal/realm/room/commands/moderation/mute"
	unmutecmd "github.com/niflaot/pixels/internal/realm/room/commands/moderation/unmute"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inmute "github.com/niflaot/pixels/networking/inbound/room/moderation/mute"
	"go.uber.org/zap"
)

// New creates a room mute packet handler.
func New(muteHandler mutecmd.Handler, unmuteHandler unmutecmd.Handler, log *zap.Logger) netconn.Handler {
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
			return unmuteDispatcher.Dispatch(context.Background(), command.Envelope[unmutecmd.Command]{Command: unmutecmd.Command{Handler: connection, RoomID: int64(payload.RoomID), PlayerID: int64(payload.PlayerID)}})
		}

		return muteDispatcher.Dispatch(context.Background(), command.Envelope[mutecmd.Command]{Command: mutecmd.Command{Handler: connection, RoomID: int64(payload.RoomID), PlayerID: int64(payload.PlayerID), Minutes: payload.Minutes}})
	}
}

// Register adds the room mute handler.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inmute.Header, handler)
}
