// Package action adapts room avatar action packets into commands.
package action

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	actioncmd "github.com/niflaot/pixels/internal/realm/room/world/commands/action"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inaction "github.com/niflaot/pixels/networking/inbound/room/entities/action"
	indance "github.com/niflaot/pixels/networking/inbound/room/entities/dance"
	inposture "github.com/niflaot/pixels/networking/inbound/room/entities/posture"
	insign "github.com/niflaot/pixels/networking/inbound/room/entities/sign"
	"go.uber.org/zap"
)

// New creates a grouped room avatar action packet handler.
func New(handler actioncmd.Handler, log *zap.Logger) func(actioncmd.Kind) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)
	return func(kind actioncmd.Kind) netconn.Handler {
		return func(connection netconn.Context, packet codec.Packet) error {
			value, err := decode(kind, packet)
			if err != nil {
				return err
			}
			return dispatcher.Dispatch(context.Background(), command.Envelope[actioncmd.Command]{
				Command:  actioncmd.Command{Handler: connection, Kind: kind, Value: value},
				Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)},
			})
		}
	}
}

// Register registers all room avatar action packets.
func Register(registry *netconn.HandlerRegistry, factory func(actioncmd.Kind) netconn.Handler) {
	_ = registry.Register(indance.Header, factory(actioncmd.KindDance))
	_ = registry.Register(inaction.Header, factory(actioncmd.KindGesture))
	_ = registry.Register(insign.Header, factory(actioncmd.KindSign))
	_ = registry.Register(inposture.Header, factory(actioncmd.KindPosture))
}

// decode decodes one action family.
func decode(kind actioncmd.Kind, packet codec.Packet) (int32, error) {
	switch kind {
	case actioncmd.KindDance:
		return indance.Decode(packet)
	case actioncmd.KindGesture:
		return inaction.Decode(packet)
	case actioncmd.KindSign:
		return insign.Decode(packet)
	case actioncmd.KindPosture:
		return inposture.Decode(packet)
	default:
		return 0, codec.ErrUnexpectedHeader
	}
}
