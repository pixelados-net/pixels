// Package relinquish adapts the room rights relinquish packet.
package relinquish

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	relinquishcmd "github.com/niflaot/pixels/internal/realm/room/commands/rights/relinquish"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inrelinquish "github.com/niflaot/pixels/networking/inbound/room/rights/relinquish"
	"go.uber.org/zap"
)

// New creates a room rights relinquish packet handler.
func New(handler relinquishcmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inrelinquish.Decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[relinquishcmd.Command]{Command: relinquishcmd.Command{Handler: connection, RoomID: int64(payload.RoomID)}})
	}
}

// Register adds the room rights relinquish handler.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inrelinquish.Header, handler)
}
