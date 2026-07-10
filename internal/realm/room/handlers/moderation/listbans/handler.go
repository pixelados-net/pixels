// Package listbans adapts the room ban-list packet.
package listbans

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	listbanscmd "github.com/niflaot/pixels/internal/realm/room/commands/moderation/listbans"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inlistbans "github.com/niflaot/pixels/networking/inbound/room/moderation/listbans"
	"go.uber.org/zap"
)

// New creates a room ban-list packet handler.
func New(handler listbanscmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inlistbans.Decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[listbanscmd.Command]{Command: listbanscmd.Command{Handler: connection, RoomID: int64(payload.RoomID)}})
	}
}

// Register adds the room ban-list handler.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inlistbans.Header, handler)
}
