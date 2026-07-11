package settings

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	settingscmd "github.com/niflaot/pixels/internal/realm/room/control/commands/settings"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	intoggle "github.com/niflaot/pixels/networking/inbound/room/mute/toggle"
	"go.uber.org/zap"
)

// NewMuteAll creates a room mute-all toggle packet handler.
func NewMuteAll(handler settingscmd.MuteAllHandler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		if err := intoggle.Decode(packet); err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[settingscmd.MuteAllCommand]{Command: settingscmd.MuteAllCommand{Handler: connection}})
	}
}

// RegisterMuteAll adds the room mute-all toggle handler.
func RegisterMuteAll(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(intoggle.Header, handler)
}
