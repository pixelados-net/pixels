// Package grant adapts the room rights grant packet.
package grant

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	grantcmd "github.com/niflaot/pixels/internal/realm/room/commands/rights/grant"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	ingrant "github.com/niflaot/pixels/networking/inbound/room/rights/grant"
	"go.uber.org/zap"
)

// New creates a room rights grant packet handler.
func New(handler grantcmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := ingrant.Decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[grantcmd.Command]{Command: grantcmd.Command{Handler: connection, PlayerID: int64(payload.PlayerID)}})
	}
}

// Register adds the room rights grant handler.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(ingrant.Header, handler)
}
