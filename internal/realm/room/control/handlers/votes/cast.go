// Package votes adapts room vote packets into commands.
package votes

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	votescmd "github.com/niflaot/pixels/internal/realm/room/control/commands/votes"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inlike "github.com/niflaot/pixels/networking/inbound/room/like"
	"go.uber.org/zap"
)

// NewCast creates a room vote packet handler.
func NewCast(handler votescmd.CastHandler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inlike.Decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[votescmd.CastCommand]{Command: votescmd.CastCommand{Handler: connection, Rating: payload.Rating}})
	}
}

// RegisterCast adds the room vote handler.
func RegisterCast(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inlike.Header, handler)
}
