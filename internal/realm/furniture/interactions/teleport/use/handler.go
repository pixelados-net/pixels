package use

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inuse "github.com/niflaot/pixels/networking/inbound/furniture/use"
	"go.uber.org/zap"
)

// New creates the furniture teleport packet adapter.
func New(handler Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inuse.Decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[Command]{
			Command: Command{Handler: connection, ItemID: int64(payload.ItemID), State: payload.State},
		})
	}
}

// Register adds the furniture teleport handler to a registry.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inuse.Header, handler)
}
