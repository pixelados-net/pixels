// Package commerce adapts catalog commerce packets to one capability command.
package commerce

import (
	"context"

	"github.com/niflaot/pixels/internal/command"
	commercecmd "github.com/niflaot/pixels/internal/realm/catalog/commands/commerce"
	"github.com/niflaot/pixels/networking/codec"
	netconn "github.com/niflaot/pixels/networking/connection"
	inbundle "github.com/niflaot/pixels/networking/inbound/catalog/bundle/discount"
	inearliest "github.com/niflaot/pixels/networking/inbound/catalog/freshness/earliest"
	inexpiration "github.com/niflaot/pixels/networking/inbound/catalog/freshness/expiration"
	inlimited "github.com/niflaot/pixels/networking/inbound/catalog/freshness/limited"
	inmark "github.com/niflaot/pixels/networking/inbound/catalog/freshness/mark"
	ingiftcheck "github.com/niflaot/pixels/networking/inbound/catalog/gift/check"
	ingiftconfig "github.com/niflaot/pixels/networking/inbound/catalog/gift/config"
	ingiftget "github.com/niflaot/pixels/networking/inbound/catalog/gift/get"
	ingiftpurchase "github.com/niflaot/pixels/networking/inbound/catalog/gift/purchase"
	invoucher "github.com/niflaot/pixels/networking/inbound/catalog/voucher/redeem"
	"go.uber.org/zap"
)

// New creates a grouped catalog commerce packet handler.
func New(handler commercecmd.Handler, log *zap.Logger) netconn.Handler {
	dispatcher, _ := command.NewDispatcher(handler)
	dispatcher.WithLogger(log)

	return func(connection netconn.Context, packet codec.Packet) error {
		input, err := decode(connection, packet)
		if err != nil {
			return err
		}

		return dispatcher.Dispatch(context.Background(), command.Envelope[commercecmd.Command]{Command: input,
			Metadata: command.Metadata{ConnectionID: string(connection.ConnectionID)}})
	}
}

// Register adds every catalog commerce header to a registry.
func Register(registry *netconn.HandlerRegistry, handler netconn.Handler) {
	for _, header := range []uint16{inbundle.Header, ingiftconfig.Header, ingiftcheck.Header, ingiftpurchase.Header,
		invoucher.Header, inmark.Header, inexpiration.Header, inearliest.Header, inlimited.Header} {
		_ = registry.Register(header, handler)
	}
	_ = registry.Register(ingiftget.Header, func(_ netconn.Context, packet codec.Packet) error {
		return ingiftget.Decode(packet)
	})
}

// decode maps one packet to a catalog commerce command.
func decode(connection netconn.Context, packet codec.Packet) (commercecmd.Command, error) {
	input := commercecmd.Command{Connection: connection}
	switch packet.Header {
	case inbundle.Header:
		input.Action = commercecmd.BundleRules
		return input, inbundle.Decode(packet)
	case ingiftconfig.Header:
		input.Action = commercecmd.GiftConfig
		return input, ingiftconfig.Decode(packet)
	case ingiftcheck.Header:
		payload, err := ingiftcheck.Decode(packet)
		input.Action, input.OfferID = commercecmd.Giftable, int64(payload.OfferID)
		return input, err
	case ingiftpurchase.Header:
		payload, err := ingiftpurchase.Decode(packet)
		input.Action, input.OfferID = commercecmd.BuyGift, int64(payload.ItemID)
		input.ReceiverName, input.Message, input.ExtraData = payload.ReceiverName, payload.GiftMessage, payload.ExtraData
		input.SpriteID, input.BoxID, input.RibbonID = payload.SpriteID, payload.BoxID, payload.RibbonID
		input.ShowMyFace = payload.ShowMyFace
		return input, err
	case invoucher.Header:
		payload, err := invoucher.Decode(packet)
		input.Action, input.Code = commercecmd.RedeemVoucher, payload.Code
		return input, err
	case inmark.Header:
		input.Action = commercecmd.MarkNew
		return input, inmark.Decode(packet)
	case inexpiration.Header:
		input.Action = commercecmd.PageExpiration
		return input, inexpiration.Decode(packet)
	case inearliest.Header:
		input.Action = commercecmd.EarliestExpiration
		return input, inearliest.Decode(packet)
	case inlimited.Header:
		input.Action = commercecmd.NextLimited
		return input, inlimited.Decode(packet)
	default:
		return input, codec.ErrUnexpectedHeader
	}
}
