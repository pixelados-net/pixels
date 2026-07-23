// Package calendar adapts campaign calendar packets to commands.
package calendar

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	calendarcmd "github.com/niflaot/pixels/internal/realm/subscription/commands/calendar"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inopen "github.com/niflaot/pixels/networking/inbound/subscription/calendar/open"
	inseasonal "github.com/niflaot/pixels/networking/inbound/subscription/calendar/seasonal"
	instaff "github.com/niflaot/pixels/networking/inbound/subscription/calendar/staff"
	"go.uber.org/zap"
)

// New creates a grouped calendar packet handler.
func New(handler calendarcmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)
	return func(connection netconn.Context, packet codec.Packet) error {
		input, err := decode(connection, packet)
		if err != nil {
			return err
		}
		return dispatcher.Dispatch(context.Background(), command.Envelope[calendarcmd.Command]{Command: input,
			Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)}})
	}
}

// Register adds calendar packet headers to a registry.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	for _, header := range []uint16{inopen.Header, instaff.Header, inseasonal.Header} {
		_ = registry.Register(header, handler)
	}
}

// decode maps one calendar packet to a command.
func decode(connection netconn.Context, packet codec.Packet) (calendarcmd.Command, error) {
	input := calendarcmd.Command{Connection: connection}
	switch packet.Header {
	case inopen.Header:
		payload, err := inopen.Decode(packet)
		input.Action, input.CampaignName, input.DayNumber = calendarcmd.Open, payload.CampaignName, payload.DayNumber
		return input, err
	case instaff.Header:
		payload, err := instaff.Decode(packet)
		input.Action, input.CampaignName, input.DayNumber = calendarcmd.OpenStaff, payload.CampaignName, payload.DayNumber
		return input, err
	case inseasonal.Header:
		input.Action = calendarcmd.Seasonal
		return input, inseasonal.Decode(packet)
	default:
		return input, codec.ErrUnexpectedHeader
	}
}
