package settings

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	settingscmd "github.com/niflaot/pixels/internal/realm/room/control/commands/settings"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inrequest "github.com/niflaot/pixels/networking/inbound/room/wordfilter/request"
	"go.uber.org/zap"
)

// NewFilterRequest creates a room word filter request packet handler.
func NewFilterRequest(handler settingscmd.FilterRequestHandler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inrequest.Decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[settingscmd.FilterRequestCommand]{Command: settingscmd.FilterRequestCommand{Handler: connection, RoomID: int64(payload.RoomID)}})
	}
}

// RegisterFilterRequest adds the room word filter request handler.
func RegisterFilterRequest(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inrequest.Header, handler)
}
