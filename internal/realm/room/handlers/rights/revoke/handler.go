// Package revoke adapts the room rights revoke packet.
package revoke

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	revokecmd "github.com/niflaot/pixels/internal/realm/room/commands/rights/revoke"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inrevoke "github.com/niflaot/pixels/networking/inbound/room/rights/revoke"
	"go.uber.org/zap"
)

// New creates a room rights revoke packet handler.
func New(handler revokecmd.Handler, log *zap.Logger) netconn.Handler {
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

		return dispatcher.Dispatch(context.Background(), command.Envelope[revokecmd.Command]{Command: revokecmd.Command{Handler: connection, PlayerIDs: playerIDs}})
	}
}

// Register adds the room rights revoke handler.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inrevoke.Header, handler)
}
