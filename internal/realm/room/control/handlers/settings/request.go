// Package settings adapts room configuration packets into commands.
package settings

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	settingscmd "github.com/niflaot/pixels/internal/realm/room/control/commands/settings"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inrequest "github.com/niflaot/pixels/networking/inbound/room/settings/request"
	"go.uber.org/zap"
)

// NewRequest creates a room settings request packet handler.
func NewRequest(handler settingscmd.RequestHandler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inrequest.Decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[settingscmd.RequestCommand]{Command: settingscmd.RequestCommand{Handler: connection, RoomID: int64(payload.RoomID)}})
	}
}

// RegisterRequest adds the room settings request handler.
func RegisterRequest(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inrequest.Header, handler)
}
