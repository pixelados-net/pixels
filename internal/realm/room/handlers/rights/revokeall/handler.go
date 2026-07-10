// Package revokeall adapts the room rights revoke-all packet.
package revokeall

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	revokeallcmd "github.com/niflaot/pixels/internal/realm/room/commands/rights/revokeall"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inrevokeall "github.com/niflaot/pixels/networking/inbound/room/rights/revokeall"
	"go.uber.org/zap"
)

// New creates a room rights revoke-all packet handler.
func New(handler revokeallcmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inrevokeall.Decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[revokeallcmd.Command]{Command: revokeallcmd.Command{Handler: connection, RoomID: int64(payload.RoomID)}})
	}
}

// Register adds the room rights revoke-all handler.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inrevokeall.Header, handler)
}
