// Package club adapts club subscription packets to commands.
package club

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	clubcmd "github.com/niflaot/pixels/internal/realm/subscription/commands/club"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inbuilders "github.com/niflaot/pixels/networking/inbound/subscription/builders/count"
	inextendbasic "github.com/niflaot/pixels/networking/inbound/subscription/club/extend/basic"
	inextendhc "github.com/niflaot/pixels/networking/inbound/subscription/club/extend/hc"
	ingiftinfo "github.com/niflaot/pixels/networking/inbound/subscription/club/gift/info"
	ingiftselect "github.com/niflaot/pixels/networking/inbound/subscription/club/gift/select"
	inoffers "github.com/niflaot/pixels/networking/inbound/subscription/club/offers"
	inpurchasebasic "github.com/niflaot/pixels/networking/inbound/subscription/club/purchase/basic"
	inpurchasevip "github.com/niflaot/pixels/networking/inbound/subscription/club/purchase/vip"
	insms "github.com/niflaot/pixels/networking/inbound/subscription/club/sms"
	inkickback "github.com/niflaot/pixels/networking/inbound/subscription/kickback"
	instatus "github.com/niflaot/pixels/networking/inbound/subscription/status"
	"go.uber.org/zap"
)

// New creates a grouped club packet handler.
func New(handler clubcmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)
	return func(connection netconn.Context, packet codec.Packet) error {
		input, err := decode(connection, packet)
		if err != nil {
			return err
		}
		return dispatcher.Dispatch(context.Background(), command.Envelope[clubcmd.Command]{Command: input,
			Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)}})
	}
}

// Register adds every club packet header to a registry.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	for _, header := range []uint16{instatus.Header, inoffers.Header, inextendbasic.Header, inextendhc.Header,
		inpurchasebasic.Header, inpurchasevip.Header, ingiftinfo.Header, ingiftselect.Header,
		inbuilders.Header, insms.Header, inkickback.Header} {
		_ = registry.Register(header, handler)
	}
}

// decode maps one club packet to a command.
func decode(connection netconn.Context, packet codec.Packet) (clubcmd.Command, error) {
	input := clubcmd.Command{Connection: connection}
	switch packet.Header {
	case instatus.Header:
		payload, err := instatus.Decode(packet)
		input.Action = clubcmd.Status
		input.ProductName = payload.ProductName
		return input, err
	case inoffers.Header:
		input.Action = clubcmd.Offers
		_, err := inoffers.Decode(packet)
		return input, err
	case inextendbasic.Header:
		input.Action = clubcmd.Extension
		return input, inextendbasic.Decode(packet)
	case inextendhc.Header:
		input.Action, input.VIP = clubcmd.Extension, true
		return input, inextendhc.Decode(packet)
	case inpurchasebasic.Header:
		payload, err := inpurchasebasic.Decode(packet)
		input.Action, input.OfferID = clubcmd.PurchaseHC, int64(payload.OfferID)
		return input, err
	case inpurchasevip.Header:
		payload, err := inpurchasevip.Decode(packet)
		input.Action, input.OfferID = clubcmd.PurchaseVIP, int64(payload.OfferID)
		return input, err
	case ingiftinfo.Header:
		input.Action = clubcmd.GiftInfo
		return input, ingiftinfo.Decode(packet)
	case ingiftselect.Header:
		payload, err := ingiftselect.Decode(packet)
		input.Action, input.GiftName = clubcmd.SelectGift, payload.GiftName
		return input, err
	case inbuilders.Header:
		input.Action = clubcmd.BuildersCount
		return input, inbuilders.Decode(packet)
	case insms.Header:
		input.Action = clubcmd.SMS
		_, err := insms.Decode(packet)
		return input, err
	case inkickback.Header:
		input.Action = clubcmd.Kickback
		return input, inkickback.Decode(packet)
	default:
		return input, codec.ErrUnexpectedHeader
	}
}
