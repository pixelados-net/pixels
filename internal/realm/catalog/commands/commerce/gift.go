package commerce

import (
	"context"
	"errors"
	"fmt"

	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	catalogprojection "github.com/niflaot/pixels/internal/realm/catalog/projection"
	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	outsoldout "github.com/niflaot/pixels/networking/outbound/catalog/limited/soldout"
	catalogoffer "github.com/niflaot/pixels/networking/outbound/catalog/offer"
	outfailed "github.com/niflaot/pixels/networking/outbound/catalog/purchase/failed"
	outok "github.com/niflaot/pixels/networking/outbound/catalog/purchase/ok"
	outunavailable "github.com/niflaot/pixels/networking/outbound/catalog/purchase/unavailable"
	outrefresh "github.com/niflaot/pixels/networking/outbound/inventory/furniture/refresh"
	outunseen "github.com/niflaot/pixels/networking/outbound/inventory/unseen"
	"go.uber.org/zap"
)

// sendGiftSuccess confirms the purchase and refreshes an online recipient.
func (handler Handler) sendGiftSuccess(ctx context.Context, input Command, result catalogservice.PurchaseResult) error {
	mapped, err := handler.mapGiftOffer(ctx, result.Item)
	if err != nil {
		return handler.sendGiftError(ctx, input, err)
	}
	packet, err := outok.Encode(mapped)
	if err := send(ctx, input, packet, err); err != nil {
		return err
	}
	if err := handler.refreshGiftRecipient(ctx, result); err != nil && handler.Log != nil {
		handler.Log.Warn("catalog gift recipient refresh failed", zap.Int64("recipient_player_id", result.RecipientPlayerID), zap.Error(err))
	}

	return nil
}

// mapGiftOffer projects the purchased offer for Nitro's success event.
func (handler Handler) mapGiftOffer(ctx context.Context, item catalogmodel.Item) (mapped catalogoffer.Offer, err error) {
	products := handler.Catalog.Products(ctx, item.ID)
	if len(products) == 0 && item.DefinitionID > 0 {
		products = []catalogmodel.Product{{DefinitionID: item.DefinitionID, Quantity: item.Amount}}
	}
	definitions := make(map[int64]furnituremodel.Definition, len(products))
	for _, product := range products {
		definition, found, findErr := handler.Catalog.Definition(ctx, product.DefinitionID)
		if findErr != nil {
			return mapped, findErr
		}
		if !found {
			return mapped, fmt.Errorf("catalog gift furniture definition %d not found", product.DefinitionID)
		}
		definitions[product.DefinitionID] = definition
	}

	return catalogprojection.OfferProducts(item, products, definitions)
}

// refreshGiftRecipient invalidates an online recipient's furniture inventory.
func (handler Handler) refreshGiftRecipient(ctx context.Context, result catalogservice.PurchaseResult) error {
	if handler.Players == nil || handler.Connections == nil {
		return nil
	}
	player, found := handler.Players.Find(result.RecipientPlayerID)
	if !found {
		return nil
	}
	peer := player.Peer()
	connection, found := handler.Connections.Get(peer.ConnectionKind(), peer.ConnectionID())
	if !found {
		return nil
	}
	itemIDs := make([]int64, len(result.GrantedItems))
	for index, item := range result.GrantedItems {
		itemIDs[index] = item.ID
	}
	packet, err := outunseen.EncodeOwned(itemIDs)
	if err != nil {
		return err
	}
	if err := connection.Send(ctx, packet); err != nil {
		return err
	}
	packet, err = outrefresh.Encode()
	if err != nil {
		return err
	}

	return connection.Send(ctx, packet)
}

// sendGiftError maps a gift purchase failure to a protocol response.
func (handler Handler) sendGiftError(ctx context.Context, input Command, purchaseErr error) error {
	if errors.Is(purchaseErr, catalogservice.ErrLimitedSoldOut) {
		packet, err := outsoldout.Encode()
		return send(ctx, input, packet, err)
	}
	if errors.Is(purchaseErr, catalogservice.ErrOfferNotGiftable) || errors.Is(purchaseErr, catalogservice.ErrOfferNotFound) ||
		errors.Is(purchaseErr, catalogservice.ErrOfferNotVisible) || errors.Is(purchaseErr, catalogservice.ErrOfferDisabled) ||
		errors.Is(purchaseErr, currencyservice.ErrInsufficientBalance) {
		packet, err := outunavailable.Encode(outunavailable.CodeIllegal)
		return send(ctx, input, packet, err)
	}
	if handler.Log != nil {
		handler.Log.Error("catalog gift purchase failed", zap.Int64("offer_id", input.OfferID), zap.Error(purchaseErr))
	}
	packet, err := outfailed.Encode(outfailed.CodeServer)

	return send(ctx, input, packet, err)
}
