package settings

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	settingscmd "github.com/niflaot/pixels/internal/realm/room/control/commands/settings"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inmodify "github.com/niflaot/pixels/networking/inbound/room/wordfilter/modify"
	"go.uber.org/zap"
)

// NewFilterModify creates a room word filter modify packet handler.
func NewFilterModify(handler settingscmd.FilterModifyHandler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		payload, err := inmodify.Decode(packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[settingscmd.FilterModifyCommand]{Command: settingscmd.FilterModifyCommand{Handler: connection, RoomID: int64(payload.RoomID), Add: payload.Add, Word: payload.Word}})
	}
}

// RegisterFilterModify adds the room word filter modify handler.
func RegisterFilterModify(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	_ = registry.Register(inmodify.Header, handler)
}
