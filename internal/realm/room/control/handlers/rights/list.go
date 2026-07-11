package rights

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	rightscmd "github.com/niflaot/pixels/internal/realm/room/control/commands/rights"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inlist "github.com/niflaot/pixels/networking/inbound/room/rights/list"
	"go.uber.org/zap"
)

// NewList creates a room rights list packet handler.
func NewList(handler rightscmd.ListHandler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inlist.Decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[rightscmd.ListCommand]{Command: rightscmd.ListCommand{Handler: connection, RoomID: int64(payload.RoomID)}})
	}
}

// RegisterList adds the room rights list handler.
func RegisterList(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inlist.Header, handler)
}
