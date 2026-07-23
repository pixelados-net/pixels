package rights

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	rightscmd "github.com/niflaot/pixels/internal/realm/room/control/commands/rights"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inrelinquish "github.com/niflaot/pixels/networking/inbound/room/rights/relinquish"
	"go.uber.org/zap"
)

// NewRelinquish creates a room rights relinquish packet handler.
func NewRelinquish(handler rightscmd.RelinquishHandler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inrelinquish.Decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[rightscmd.RelinquishCommand]{Command: rightscmd.RelinquishCommand{Handler: connection, RoomID: int64(payload.RoomID)}})
	}
}

// RegisterRelinquish adds the room rights relinquish handler.
func RegisterRelinquish(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inrelinquish.Header, handler)
}
