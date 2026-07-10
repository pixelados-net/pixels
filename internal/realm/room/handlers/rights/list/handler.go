// Package list adapts the room rights list packet.
package list

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	listcmd "github.com/niflaot/pixels/internal/realm/room/commands/rights/list"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inlist "github.com/niflaot/pixels/networking/inbound/room/rights/list"
	"go.uber.org/zap"
)

// New creates a room rights list packet handler.
func New(handler listcmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inlist.Decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[listcmd.Command]{Command: listcmd.Command{Handler: connection, RoomID: int64(payload.RoomID)}})
	}
}

// Register adds the room rights list handler.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inlist.Header, handler)
}
