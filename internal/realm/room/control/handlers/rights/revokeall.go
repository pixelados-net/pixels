package rights

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	rightscmd "github.com/niflaot/pixels/internal/realm/room/control/commands/rights"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inrevokeall "github.com/niflaot/pixels/networking/inbound/room/rights/revokeall"
	"go.uber.org/zap"
)

// NewRevokeAll creates a room rights revoke-all packet handler.
func NewRevokeAll(handler rightscmd.RevokeAllHandler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inrevokeall.Decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[rightscmd.RevokeAllCommand]{Command: rightscmd.RevokeAllCommand{Handler: connection, RoomID: int64(payload.RoomID)}})
	}
}

// RegisterRevokeAll adds the room rights revoke-all handler.
func RegisterRevokeAll(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inrevokeall.Header, handler)
}
