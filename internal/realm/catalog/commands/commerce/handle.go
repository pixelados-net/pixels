package commerce

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/niflaot/pixels/internal/command"
	catalogsession "github.com/niflaot/pixels/internal/realm/catalog/commands/session"
	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	"github.com/niflaot/pixels/networking/codec"
	outbundle "github.com/niflaot/pixels/networking/outbound/catalog/bundle/discount"
	outearliest "github.com/niflaot/pixels/networking/outbound/catalog/freshness/earliest"
	outexpiration "github.com/niflaot/pixels/networking/outbound/catalog/freshness/expiration"
	outlimited "github.com/niflaot/pixels/networking/outbound/catalog/freshness/limited"
	outconfig "github.com/niflaot/pixels/networking/outbound/catalog/gift/config"
	outgiftable "github.com/niflaot/pixels/networking/outbound/catalog/gift/giftable"
	outnotfound "github.com/niflaot/pixels/networking/outbound/catalog/gift/notfound"
	outfailed "github.com/niflaot/pixels/networking/outbound/catalog/voucher/failed"
	outok "github.com/niflaot/pixels/networking/outbound/catalog/voucher/ok"
)

// Handle executes one catalog commerce command.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	if envelope.Command.Action == BundleRules {
		packet, err := outbundle.Encode()
		return send(ctx, envelope.Command, packet, err)
	}
	if envelope.Command.Action == GiftConfig {
		options := handler.GiftOptions
		packet, err := outconfig.Encode(outconfig.Options{Price: options.Price, Wrappers: options.Wrappers, Boxes: options.Boxes,
			Ribbons: options.Ribbons, DefaultGifts: options.DefaultGifts})
		return send(ctx, envelope.Command, packet, err)
	}
	player, err := catalogsession.Player(envelope.Command.Connection, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	switch envelope.Command.Action {
	case Giftable:
		item, found := handler.Catalog.Item(envelope.Command.OfferID)
		packet, encodeErr := outgiftable.Encode(int32(envelope.Command.OfferID), found && item.Giftable)
		return send(ctx, envelope.Command, packet, encodeErr)
	case BuyGift:
		boxID, ribbonID, valid := handler.GiftOptions.Resolve(envelope.Command.SpriteID, envelope.Command.BoxID, envelope.Command.RibbonID)
		if !valid {
			return handler.sendGiftError(ctx, envelope.Command, catalogservice.ErrOfferNotGiftable)
		}
		result, purchaseErr := handler.Catalog.PurchaseGift(ctx, catalogservice.GiftPurchaseParams{BuyerID: player.ID(), ReceiverName: envelope.Command.ReceiverName, CatalogItemID: envelope.Command.OfferID, HasClub: catalogsession.HasClub(player), SpriteID: envelope.Command.SpriteID, BoxID: boxID, RibbonID: ribbonID, Message: envelope.Command.Message, ExtraData: envelope.Command.ExtraData, ShowMyFace: envelope.Command.ShowMyFace})
		if errors.Is(purchaseErr, catalogservice.ErrGiftReceiverNotFound) {
			packet, encodeErr := outnotfound.Encode(envelope.Command.ReceiverName)
			return send(ctx, envelope.Command, packet, encodeErr)
		}
		if purchaseErr != nil {
			return handler.sendGiftError(ctx, envelope.Command, purchaseErr)
		}
		return handler.sendGiftSuccess(ctx, envelope.Command, result)
	case RedeemVoucher:
		result, redeemErr := handler.Catalog.RedeemVoucher(ctx, player.ID(), envelope.Command.Code)
		if redeemErr == nil {
			packet, encodeErr := outok.Encode(result.ProductCode)
			return send(ctx, envelope.Command, packet, encodeErr)
		}
		code := outfailed.Invalid
		if errors.Is(redeemErr, catalogservice.ErrVoucherAlreadyUsed) {
			code = outfailed.AlreadyUsed
		}
		if errors.Is(redeemErr, catalogservice.ErrVoucherExhausted) {
			code = outfailed.Expired
		}
		packet, encodeErr := outfailed.Encode(code)
		return send(ctx, envelope.Command, packet, encodeErr)
	case MarkNew:
		return handler.Catalog.MarkNewAdditionsSeen(ctx, player.ID())
	case PageExpiration:
		pageID, found := player.OpenCatalog().CurrentPage()
		page, seconds, exists := handler.Catalog.PageExpiration(pageID, time.Now())
		if !found || !exists {
			packet, encodeErr := outexpiration.Encode(0, 0, "")
			return send(ctx, envelope.Command, packet, encodeErr)
		}
		packet, encodeErr := outexpiration.Encode(int32(page.ID), seconds, page.Name)
		return send(ctx, envelope.Command, packet, encodeErr)
	case EarliestExpiration:
		pages, readErr := handler.Catalog.Pages(ctx, player.ID(), catalogsession.HasClub(player))
		if readErr != nil {
			return readErr
		}
		page, seconds, found := handler.Catalog.EarliestExpiration(pages, time.Now())
		if !found {
			packet, encodeErr := outearliest.Encode("", 0)
			return send(ctx, envelope.Command, packet, encodeErr)
		}
		packet, encodeErr := outearliest.Encode(page.Name, seconds)
		return send(ctx, envelope.Command, packet, encodeErr)
	case NextLimited:
		item, seconds, found := handler.Catalog.NextLimited(time.Now())
		if !found {
			packet, encodeErr := outlimited.Encode(math.MaxInt32, 0, 0, "")
			return send(ctx, envelope.Command, packet, encodeErr)
		}
		definition, _, readErr := handler.Catalog.Definition(ctx, item.DefinitionID)
		if readErr != nil {
			return readErr
		}
		packet, encodeErr := outlimited.Encode(seconds, int32(item.PageID), int32(item.ID), string(definition.Kind))
		return send(ctx, envelope.Command, packet, encodeErr)
	default:
		return nil
	}
}

// send sends an encoded commerce response.
func send(ctx context.Context, input Command, packet codec.Packet, err error) error {
	if err != nil {
		return err
	}

	return input.Connection.Send(ctx, packet)
}
