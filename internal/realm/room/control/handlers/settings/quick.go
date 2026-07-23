package settings

import (
	"context"
	"github.com/niflaot/pixels/internal/command"
	settingscmd "github.com/niflaot/pixels/internal/realm/room/control/commands/settings"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inquick "github.com/niflaot/pixels/networking/inbound/room/settings/quick"
	"go.uber.org/zap"
)

// NewQuick creates the focused room settings packet handler.
func NewQuick(handler settingscmd.QuickHandler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)
	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inquick.Decode(packet)
		if err != nil {
			return err
		}
		return dispatcher.Dispatch(context.Background(), command.Envelope[settingscmd.QuickCommand]{Command: settingscmd.QuickCommand{Handler: connection, RoomID: int64(payload.RoomID), CategoryID: int64(payload.CategoryID), TradeMode: payload.TradeMode}})
	}
}

// RegisterQuick adds the focused room settings handler.
func RegisterQuick(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inquick.Header, handler)
}
