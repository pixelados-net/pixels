// Package model contains the room model packet handler.
package model

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	modelcmd "github.com/niflaot/pixels/internal/realm/room/commands/model"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inmodel "github.com/niflaot/pixels/networking/inbound/room/model"
	"go.uber.org/zap"
)

// New creates a room model packet handler.
func New(handler modelcmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		if _, err := inmodel.Decode(packet); err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[modelcmd.Command]{
			Command:  modelcmd.Command{Handler: connection},
			Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)},
		})
	}
}

// Register adds the room model handler to a registry.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inmodel.Header, handler)
}
