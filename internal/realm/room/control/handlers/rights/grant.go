// Package rights adapts room rights packets into commands.
package rights

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	rightscmd "github.com/niflaot/pixels/internal/realm/room/control/commands/rights"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	ingrant "github.com/niflaot/pixels/networking/inbound/room/rights/grant"
	"go.uber.org/zap"
)

// NewGrant creates a room rights grant packet handler.
func NewGrant(handler rightscmd.GrantHandler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := ingrant.Decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[rightscmd.GrantCommand]{Command: rightscmd.GrantCommand{Handler: connection, PlayerID: int64(payload.PlayerID)}})
	}
}

// RegisterGrant adds the room rights grant handler.
func RegisterGrant(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(ingrant.Header, handler)
}
