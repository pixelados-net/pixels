// Package settings adapts GET_SOUND_SETTINGS to a chat settings command.
package settings

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	settingscmd "github.com/niflaot/pixels/internal/realm/chat/commands/settings"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	insettings "github.com/niflaot/pixels/networking/inbound/chat/settings"
	"go.uber.org/zap"
)

// New creates a GET_SOUND_SETTINGS packet handler.
func New(handler settingscmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		if err := insettings.Decode(packet); err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[settingscmd.Command]{
			Command:  settingscmd.Command{Handler: connection},
			Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)},
		})
	}
}

// Register adds GET_SOUND_SETTINGS to a handler registry.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(insettings.Header, handler)
}
