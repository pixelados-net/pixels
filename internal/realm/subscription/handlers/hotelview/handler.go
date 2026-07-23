// Package hotelview adapts hotel-view packets to subscription commands.
package hotelview

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	hotelviewcmd "github.com/niflaot/pixels/internal/realm/subscription/commands/hotelview"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inbonus "github.com/niflaot/pixels/networking/inbound/subscription/bonusrare/request"
	incampaign "github.com/niflaot/pixels/networking/inbound/subscription/campaign/start"
	incountdown "github.com/niflaot/pixels/networking/inbound/subscription/countdown"
	"go.uber.org/zap"
)

// New creates a grouped hotel-view packet handler.
func New(handler hotelviewcmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)
	return func(connection netconn.Context, packet codec.Packet) error {
		input, err := decode(connection, packet)
		if err != nil {
			return err
		}
		return dispatcher.Dispatch(context.Background(), command.Envelope[hotelviewcmd.Command]{Command: input,
			Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)}})
	}
}

// Register adds hotel-view packet headers to a registry.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	if registry == nil || handler == nil {
		return
	}
	for _, header := range []uint16{inbonus.Header, incountdown.Header, incampaign.Header} {
		_ = registry.Register(header, handler)
	}
}

// decode maps one hotel-view packet to a command.
func decode(connection netconn.Context, packet codec.Packet) (hotelviewcmd.Command, error) {
	input := hotelviewcmd.Command{Connection: connection}
	switch packet.Header {
	case inbonus.Header:
		input.Action = hotelviewcmd.BonusRare
		return input, inbonus.Decode(packet)
	case incountdown.Header:
		input.Action = hotelviewcmd.Countdown
		value, err := incountdown.Decode(packet)
		input.Value = value
		return input, err
	case incampaign.Header:
		input.Action = hotelviewcmd.StartCampaign
		value, err := incampaign.Decode(packet)
		input.Value = value
		return input, err
	default:
		return input, codec.ErrUnexpectedHeader
	}
}
