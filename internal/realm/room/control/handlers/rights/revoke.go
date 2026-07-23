package rights

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	rightscmd "github.com/niflaot/pixels/internal/realm/room/control/commands/rights"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inrevoke "github.com/niflaot/pixels/networking/inbound/room/rights/revoke"
	"go.uber.org/zap"
)

// NewRevoke creates a room rights revoke packet handler.
func NewRevoke(handler rightscmd.RevokeHandler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inrevoke.Decode(packet)
		if err != nil {
			return err
		}
		playerIDs := make([]int64, len(payload.PlayerIDs))
		for index := range payload.PlayerIDs {
			playerIDs[index] = int64(payload.PlayerIDs[index])
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[rightscmd.RevokeCommand]{Command: rightscmd.RevokeCommand{Handler: connection, PlayerIDs: playerIDs}})
	}
}

// RegisterRevoke adds the room rights revoke handler.
func RegisterRevoke(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inrevoke.Header, handler)
}
