package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	cataloggiftevent "github.com/niflaot/pixels/internal/realm/catalog/events/gift"
	catalogvoucherevent "github.com/niflaot/pixels/internal/realm/catalog/events/voucher"
	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	playereffect "github.com/niflaot/pixels/internal/realm/player/effect"
	"github.com/niflaot/pixels/pkg/bus"
)

// finishRewards completes dependent furniture and avatar effect grants.
func (service *Service) finishRewards(ctx context.Context, playerID int64, item catalogmodel.Item, result *PurchaseResult) error {
	if err := service.pairGrantedTeleports(ctx, playerID, item, result.GrantedItems); err != nil {
		return err
	}
	if item.GrantsEffectID == nil {
		return nil
	}
	if service.effects == nil {
		return ErrCommerceUnavailable
	}
	if _, err := service.effects.GrantEnabled(ctx, playerID, *item.GrantsEffectID, item.GrantsEffectDurationSeconds, playereffect.SourceCatalog); err != nil {
		return fmt.Errorf("grant and enable catalog effect %d: %w", *item.GrantsEffectID, err)
	}
	result.GrantedEffectID = item.GrantsEffectID

	return nil
}

// validateAmount validates anti-cheat purchase quantity rules.
func validateAmount(item catalogmodel.Item, products []catalogmodel.Product, amount int32, overrideQuantity bool) error {
	if item.IsRoomBundle() && amount != 1 {
		return ErrInvalidAmount
	}
	if item.IsRoomBundle() {
		return nil
	}
	if amount <= 0 || item.IsLimited() && amount != 1 {
		return ErrInvalidAmount
	}
	if amount > 1 && !overrideQuantity && !item.BulkDiscountEligible(len(products) > 0) {
		return ErrInvalidAmount
	}
	units := item.Amount
	if len(products) > 0 {
		units = 0
		for _, product := range products {
			units += product.Quantity
		}
	}
	if int64(units)*int64(amount) > int64(MaxPurchaseUnits) {
		return ErrInvalidAmount
	}
	return nil
}

// pairGrantedTeleports pairs every adjacent teleport instance from one offer.
func (service *Service) pairGrantedTeleports(ctx context.Context, playerID int64, item catalogmodel.Item, granted []furnituremodel.Item) error {
	definition, found := service.cache.definition(item.DefinitionID)
	if !found || (definition.InteractionType != "teleport" && definition.InteractionType != "teleport_tile") {
		return nil
	}
	if service.teleportPairs == nil || len(granted) == 0 || len(granted)%2 != 0 {
		return ErrTeleportPairing
	}
	for index := 0; index < len(granted); index += 2 {
		if err := service.teleportPairs.PairTeleports(ctx, playerID, granted[index].ID, granted[index+1].ID); err != nil {
			return fmt.Errorf("%w: %v", ErrTeleportPairing, err)
		}
	}

	return nil
}

// charge deducts the configured offer price.
func (service *Service) charge(ctx context.Context, playerID int64, item catalogmodel.Item, params PurchaseParams) (int64, int64, int64, error) {
	currencyType, cost := item.PointsType, item.CostPoints
	if item.IsCredits() {
		currencyType, cost = catalogmodel.CreditsType, item.CostCredits
	}
	if params.OverrideCredits != nil {
		currencyType, cost = catalogmodel.CreditsType, *params.OverrideCredits
	}
	if params.OverridePoints != nil {
		cost = *params.OverridePoints
		if params.OverridePointsType != nil {
			currencyType = *params.OverridePointsType
		}
	}
	if cost == 0 || params.Free {
		return 0, 0, 0, nil
	}
	payable := params.Amount
	if params.Amount > 1 && params.OverrideCredits == nil && params.OverridePoints == nil {
		payable -= DiscountedUnits(params.Amount)
	}
	cost *= int64(payable)
	balance, err := service.currencies.Grant(ctx, currencyservice.GrantParams{PlayerID: playerID, CurrencyType: currencyType,
		Amount: -cost, Reason: "catalog_purchase", ActorKind: currencyservice.ActorPlayer})
	if currencyType == catalogmodel.CreditsType {
		return balance, cost, 0, err
	}

	return balance, 0, cost, err
}

// PurchaseGift buys one wrapped offer for another player.
func (service *Service) PurchaseGift(ctx context.Context, params GiftPurchaseParams) (PurchaseResult, error) {
	if service.players == nil || len(params.Message) > 255 {
		return PurchaseResult{}, ErrGiftReceiverNotFound
	}
	receiver, found, err := service.players.FindByUsername(ctx, strings.TrimSpace(params.ReceiverName))
	if err != nil || !found {
		if err != nil {
			return PurchaseResult{}, err
		}
		return PurchaseResult{}, ErrGiftReceiverNotFound
	}
	item, found := service.cache.item(params.CatalogItemID)
	if !found || !item.Giftable || item.IsRoomBundle() {
		return PurchaseResult{}, ErrOfferNotGiftable
	}
	var senderID *int64
	if params.ShowMyFace {
		senderID = &params.BuyerID
	}
	result, err := service.Purchase(ctx, PurchaseParams{PlayerID: params.BuyerID, RecipientPlayerID: receiver.Player.ID, CatalogItemID: params.CatalogItemID, HasClub: params.HasClub, Amount: 1, ExtraData: params.ExtraData, Gift: &GiftMetadata{SpriteID: params.SpriteID, BoxID: params.BoxID, RibbonID: params.RibbonID, SenderPlayerID: senderID, Message: params.Message}})
	result.RecipientPlayerID = receiver.Player.ID
	if err == nil && service.events != nil {
		_ = service.events.Publish(ctx, bus.Event{Name: cataloggiftevent.Name, Payload: cataloggiftevent.Payload{BuyerID: params.BuyerID, ReceiverID: receiver.Player.ID, CatalogItemID: params.CatalogItemID}})
	}

	return result, err
}

// RedeemVoucher redeems one case-insensitive voucher exactly once per player.
func (service *Service) RedeemVoucher(ctx context.Context, playerID int64, code string) (RedeemResult, error) {
	if service.commerce == nil {
		return RedeemResult{}, ErrCommerceUnavailable
	}
	voucher, found, err := service.commerce.FindVoucherByCode(ctx, strings.TrimSpace(code))
	if err != nil {
		return RedeemResult{}, err
	}
	if !found || !voucher.Enabled || voucher.ExpiresAt != nil && !time.Now().Before(*voucher.ExpiresAt) {
		return RedeemResult{}, ErrVoucherInvalid
	}
	count, err := service.commerce.CountVoucherRedemptions(ctx, voucher.ID)
	if err != nil {
		return RedeemResult{}, err
	}
	if voucher.RedemptionCap != nil && count >= *voucher.RedemptionCap {
		return RedeemResult{}, ErrVoucherExhausted
	}
	result := RedeemResult{ProductCode: voucher.Code}
	err = service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		if insertErr := service.commerce.InsertVoucherRedemption(txCtx, voucher.ID, playerID); insertErr != nil {
			return ErrVoucherAlreadyUsed
		}
		if voucher.CostCredits != 0 {
			if _, grantErr := service.currencies.Grant(txCtx, currencyservice.GrantParams{PlayerID: playerID, CurrencyType: catalogmodel.CreditsType, Amount: voucher.CostCredits, Reason: "catalog_voucher", ActorKind: currencyservice.ActorSystem}); grantErr != nil {
				return grantErr
			}
		}
		if voucher.CostPoints != 0 {
			if _, grantErr := service.currencies.Grant(txCtx, currencyservice.GrantParams{PlayerID: playerID, CurrencyType: voucher.PointsType, Amount: voucher.CostPoints, Reason: "catalog_voucher", ActorKind: currencyservice.ActorSystem}); grantErr != nil {
				return grantErr
			}
		}
		if voucher.CatalogItemID == nil {
			return nil
		}
		item, found := service.cache.item(*voucher.CatalogItemID)
		if !found || !item.Enabled {
			return ErrOfferNotFound
		}
		purchase := PurchaseResult{Item: item}
		purchaseErr := service.commitPurchase(txCtx, PurchaseParams{PlayerID: playerID, CatalogItemID: item.ID, HasClub: true, Amount: 1, Free: true}, item, service.cache.products(item.ID), &purchase)
		result.GrantedItems = purchase.GrantedItems
		return purchaseErr
	})
	if err == nil && service.events != nil {
		_ = service.events.Publish(ctx, bus.Event{Name: catalogvoucherevent.Name, Payload: catalogvoucherevent.Payload{PlayerID: playerID, VoucherID: voucher.ID}})
	}

	return result, err
}
