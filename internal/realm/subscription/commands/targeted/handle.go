package targeted

import (
	"context"
	"time"

	"github.com/niflaot/pixels/internal/command"
	catalogsession "github.com/niflaot/pixels/internal/realm/catalog/commands/session"
	"github.com/niflaot/pixels/internal/realm/subscription/record"
	"github.com/niflaot/pixels/networking/codec"
	outpurchase "github.com/niflaot/pixels/networking/outbound/catalog/purchase/ok"
	outrefresh "github.com/niflaot/pixels/networking/outbound/inventory/furniture/refresh"
	outunseen "github.com/niflaot/pixels/networking/outbound/inventory/unseen"
	outnotfound "github.com/niflaot/pixels/networking/outbound/subscription/targeted/notfound"
	outtarget "github.com/niflaot/pixels/networking/outbound/subscription/targeted/offer"
	outproduct "github.com/niflaot/pixels/networking/outbound/subscription/targeted/product"
	"github.com/niflaot/pixels/pkg/i18n"
)

// Handle executes one targeted-offer command.
func (handler Handler) Handle(ctx context.Context, envelope command.Envelope[Command]) error {
	player, err := catalogsession.Player(envelope.Command.Connection, handler.Bindings, handler.Players)
	if err != nil {
		return err
	}
	switch envelope.Command.Action {
	case Current, Next:
		after := int64(0)
		if envelope.Command.Action == Next {
			after = envelope.Command.OfferID
		}
		offer, found, readErr := handler.Subscriptions.TargetedOffer(ctx, player.ID(), after)
		if readErr != nil {
			return readErr
		}
		if !found {
			packet, encodeErr := outnotfound.Encode()
			return send(ctx, envelope.Command.Connection, packet, encodeErr)
		}
		return handler.sendOffer(ctx, envelope.Command.Connection, offer)
	case Purchase:
		result, purchaseErr := handler.Subscriptions.PurchaseTargetedOffer(ctx, player.ID(), envelope.Command.OfferID, envelope.Command.Quantity)
		if purchaseErr != nil {
			return purchaseErr
		}
		mapped, mapErr := handler.mapCatalogOffer(ctx, result.Item)
		if mapErr != nil {
			return mapErr
		}
		itemIDs := make([]int64, 0, len(result.GrantedItems))
		for _, item := range result.GrantedItems {
			itemIDs = append(itemIDs, item.ID)
		}
		packet, encodeErr := outunseen.EncodeOwned(itemIDs)
		if err := send(ctx, envelope.Command.Connection, packet, encodeErr); err != nil {
			return err
		}
		packet, encodeErr = outpurchase.Encode(mapped)
		if err := send(ctx, envelope.Command.Connection, packet, encodeErr); err != nil {
			return err
		}
		packet, encodeErr = outrefresh.Encode()
		return send(ctx, envelope.Command.Connection, packet, encodeErr)
	case State:
		return handler.Subscriptions.SetTargetedState(ctx, player.ID(), envelope.Command.OfferID, envelope.Command.Dismissed)
	case Product:
		item, found := handler.Catalog.Item(envelope.Command.OfferID)
		if !found {
			packet, encodeErr := outnotfound.Encode()
			return send(ctx, envelope.Command.Connection, packet, encodeErr)
		}
		mapped, mapErr := handler.mapCatalogOffer(ctx, item)
		packet, encodeErr := outproduct.Encode(mapped)
		if mapErr != nil {
			return mapErr
		}
		return send(ctx, envelope.Command.Connection, packet, encodeErr)
	default:
		return nil
	}
}

// sendOffer sends one localized targeted offer.
func (handler Handler) sendOffer(ctx context.Context, connection interface {
	Send(context.Context, codec.Packet) error
}, offer record.TargetedOffer) error {
	item, found := handler.Catalog.Item(offer.CatalogItemID)
	if !found {
		packet, err := outnotfound.Encode()
		if err != nil {
			return err
		}
		return connection.Send(ctx, packet)
	}
	seconds := int32(0)
	if offer.ExpiresAt != nil && offer.ExpiresAt.After(time.Now()) {
		seconds = int32(offer.ExpiresAt.Sub(time.Now()) / time.Second)
	}
	title, description := offer.TitleKey, offer.DescriptionKey
	if handler.Translations != nil {
		title = handler.Translations.Default(i18n.Key(offer.TitleKey))
		description = handler.Translations.Default(i18n.Key(offer.DescriptionKey))
	}
	packet, err := outtarget.Encode(outtarget.Offer{ID: int32(offer.ID), Identifier: item.Name,
		ProductCode: item.Name, PriceCredits: clampInt32(offer.PriceCredits), PricePoints: clampInt32(offer.PricePoints),
		PointsType: offer.PointsType, PurchaseLimit: offer.PurchaseLimit - offer.PurchasesCount, ExpirationSeconds: seconds,
		Title: title, Description: description, ImageURL: offer.ImageURL, IconURL: offer.IconURL})
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
}
