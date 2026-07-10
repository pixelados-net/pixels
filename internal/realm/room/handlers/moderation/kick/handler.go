// Package kick adapts the room kick packet.
package kick

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	kickcmd "github.com/niflaot/pixels/internal/realm/room/commands/moderation/kick"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inkick "github.com/niflaot/pixels/networking/inbound/room/moderation/kick"
	"go.uber.org/zap"
)

// New creates a room kick packet handler.
func New(handler kickcmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inkick.Decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[kickcmd.Command]{Command: kickcmd.Command{Handler: connection, PlayerID: int64(payload.PlayerID)}})
	}
}

// Register adds the room kick handler.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inkick.Header, handler)
}
