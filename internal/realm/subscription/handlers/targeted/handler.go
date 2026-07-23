// Package targeted adapts personalized offer packets to commands.
package targeted

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	targetedcmd "github.com/niflaot/pixels/internal/realm/subscription/commands/targeted"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inproduct "github.com/niflaot/pixels/networking/inbound/subscription/product"
	incurrent "github.com/niflaot/pixels/networking/inbound/subscription/targeted/current"
	innext "github.com/niflaot/pixels/networking/inbound/subscription/targeted/next"
	inpurchase "github.com/niflaot/pixels/networking/inbound/subscription/targeted/purchase"
	instate "github.com/niflaot/pixels/networking/inbound/subscription/targeted/state"
	inviewed "github.com/niflaot/pixels/networking/inbound/subscription/targeted/viewed"
	"go.uber.org/zap"
)

// New creates a grouped targeted-offer handler.
func New(handler targetedcmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)
	return func(connection netconn.Context, packet codec.Packet) error {
		input, err := decode(connection, packet)
		if err != nil {
			return err
		}
		return dispatcher.Dispatch(context.Background(), command.Envelope[targetedcmd.Command]{Command: input,
			Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)}})
	}
}

// Register adds targeted-offer headers to a registry.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	for _, header := range []uint16{incurrent.Header, innext.Header, inpurchase.Header, instate.Header, inviewed.Header, inproduct.Header} {
		_ = registry.Register(header, handler)
	}
}

// decode maps one targeted-offer packet to a command.
func decode(connection netconn.Context, packet codec.Packet) (targetedcmd.Command, error) {
	input := targetedcmd.Command{Connection: connection}
	switch packet.Header {
	case incurrent.Header:
		input.Action = targetedcmd.Current
		return input, incurrent.Decode(packet)
	case innext.Header:
		payload, err := innext.Decode(packet)
		input.Action, input.OfferID = targetedcmd.Next, int64(payload.OfferID)
		return input, err
	case inpurchase.Header:
		payload, err := inpurchase.Decode(packet)
		input.Action, input.OfferID, input.Quantity = targetedcmd.Purchase, int64(payload.OfferID), payload.Quantity
		return input, err
	case instate.Header:
		payload, err := instate.Decode(packet)
		input.Action, input.OfferID, input.Dismissed = targetedcmd.State, int64(payload.OfferID), payload.State != 0
		return input, err
	case inviewed.Header:
		payload, err := inviewed.Decode(packet)
		input.Action, input.OfferID = targetedcmd.State, int64(payload.OfferID)
		return input, err
	case inproduct.Header:
		payload, err := inproduct.Decode(packet)
		input.Action, input.OfferID = targetedcmd.Product, int64(payload.OfferID)
		return input, err
	default:
		return input, codec.ErrUnexpectedHeader
	}
}
