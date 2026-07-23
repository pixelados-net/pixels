// Package handitem adapts carried-item packets to room commands.
package handitem

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	handcmd "github.com/niflaot/pixels/internal/realm/room/world/commands/handitem"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	indrop "github.com/niflaot/pixels/networking/inbound/room/entities/handitem/drop"
	ingive "github.com/niflaot/pixels/networking/inbound/room/entities/handitem/give"
	"go.uber.org/zap"
)

// Register registers drop and give packet adapters.
func Register(registry *netconn.HandlerRegistry, handler handcmd.Handler, log *zap.Logger) {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)
	_ = registry.Register(indrop.Header, func(connection netconn.Context, packet codec.Packet) error {
		if _, err := indrop.Decode(packet); err != nil {
			return err
		}
		return dispatch(dispatcher, connection, handcmd.Command{Handler: connection, Kind: handcmd.KindDrop})
	})
	_ = registry.Register(ingive.Header, func(connection netconn.Context, packet codec.Packet) error {
		payload, err := ingive.Decode(packet)
		if err != nil {
			return err
		}
		return dispatch(dispatcher, connection, handcmd.Command{Handler: connection, Kind: handcmd.KindGive, TargetUnitID: int64(payload.UnitID)})
	})
}

// dispatch sends one carried-item command with connection metadata.
func dispatch(dispatcher *command.Dispatcher[handcmd.Command], connection netconn.Context, value handcmd.Command) error {
	return dispatcher.Dispatch(context.Background(), command.Envelope[handcmd.Command]{Command: value, Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)}})
}
