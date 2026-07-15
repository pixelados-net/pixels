package service

import (
	"context"
	"fmt"

	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	roombundle "github.com/niflaot/pixels/internal/realm/room/record/bundle"
)

const (
	// MaxPurchaseUnits is the maximum number of furniture instances in one purchase.
	MaxPurchaseUnits int32 = 100
	// DiscountBatchSize controls the base bulk discount cadence.
	DiscountBatchSize int32 = 6
)

var (
	// AdditionalDiscountThresholds stores extra free-unit thresholds.
	AdditionalDiscountThresholds = [...]int32{40, 99}
)

// Purchase buys one catalog offer.
func (service *Service) Purchase(ctx context.Context, params PurchaseParams) (PurchaseResult, error) {
	if params.Amount == 0 {
		params.Amount = 1
	}
	item, err := service.purchaseOffer(ctx, params)
	if err != nil {
		return PurchaseResult{}, err
	}

	products := service.cache.products(item.ID)
	if item.IsRoomBundle() {
		if service.roomBundles == nil {
			return PurchaseResult{}, ErrCommerceUnavailable
		}
		products, err = service.roomBundleProducts(ctx, item)
		if err != nil {
			return PurchaseResult{}, fmt.Errorf("preview room bundle %d: %w", item.ID, err)
		}
		if len(products) == 0 {
			return PurchaseResult{}, ErrOfferDisabled
		}
	}
	if item.GrantsEffectID != nil && (params.Gift != nil || params.RecipientPlayerID != 0) {
		return PurchaseResult{}, ErrOfferNotGiftable
	}
	overrideQuantity := params.OverrideCredits != nil || params.OverridePoints != nil
	if err := validateAmount(item, products, params.Amount, overrideQuantity); err != nil {
		return PurchaseResult{}, err
	}
	result := PurchaseResult{Item: item, Products: products}
	err = service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		return service.commitPurchase(txCtx, params, item, products, &result)
	})
	if err != nil {
		return PurchaseResult{}, err
	}

	service.refreshAfterLimited(ctx, item)
	service.publishPurchase(ctx, params.PlayerID, result)

	return result, nil
}

// purchaseOffer validates and resolves one cached offer and page.
func (service *Service) purchaseOffer(ctx context.Context, params PurchaseParams) (catalogmodel.Item, error) {
	if params.PlayerID <= 0 {
		return catalogmodel.Item{}, ErrInvalidPlayerID
	}
	if params.CatalogItemID <= 0 {
		return catalogmodel.Item{}, ErrInvalidOfferID
	}

	item, found := service.cache.item(params.CatalogItemID)
	if !found {
		return catalogmodel.Item{}, ErrOfferNotFound
	}
	if !item.Enabled {
		return catalogmodel.Item{}, ErrOfferDisabled
	}
	page, found := service.cache.page(item.PageID)
	if !found {
		return catalogmodel.Item{}, ErrPageNotFound
	}
	accessible, err := service.pageAccessible(ctx, page, params.PlayerID, params.HasClub)
	if err != nil {
		return catalogmodel.Item{}, err
	}
	if !accessible || (item.ClubOnly && !params.HasClub) {
		return catalogmodel.Item{}, ErrOfferNotVisible
	}

	return item, nil
}

// commitPurchase charges and grants one offer inside the active transaction.
func (service *Service) commitPurchase(ctx context.Context, params PurchaseParams, item catalogmodel.Item, products []catalogmodel.Product, result *PurchaseResult) error {
	if item.IsLimited() {
		number, reserved, err := service.store.ReserveLimitedUnit(ctx, item.ID, params.PlayerID)
		if err != nil {
			return err
		}
		if !reserved {
			return ErrLimitedSoldOut
		}
		result.LimitedUnitNumber = &number
	}

	if item.IsRoomBundle() {
		if service.roomBundles == nil || service.players == nil || params.Gift != nil || params.RecipientPlayerID != 0 {
			return ErrOfferNotGiftable
		}
		buyer, found, err := service.players.FindByID(ctx, params.PlayerID)
		if err != nil {
			return err
		}
		if !found {
			return ErrInvalidPlayerID
		}
		created, err := service.roomBundles.Clone(ctx, roombundle.CloneParams{TemplateRoomID: *item.RoomBundleTemplateRoomID, BuyerPlayerID: params.PlayerID, BuyerName: buyer.Player.Username, CatalogItemID: item.ID})
		if err != nil {
			return err
		}
		result.CreatedRoomID = &created.Room.ID
		result.CreatedRoomName = created.Room.Name
		result.ClonedFurnitureCount = created.FurnitureCount
		result.ClonedBotCount = created.BotCount
	}

	balance, credits, points, err := service.charge(ctx, params.PlayerID, item, params)
	if err != nil {
		return err
	}
	result.ChargedCredits = credits
	result.ChargedPoints = points
	if item.IsCredits() {
		result.NewCreditsBalance = balance
	} else {
		result.NewPointsBalance = balance
	}

	if item.IsRoomBundle() {
		return service.logPurchase(ctx, params, item, result, credits, points)
	}
	recipientID := params.RecipientPlayerID
	if recipientID == 0 {
		recipientID = params.PlayerID
	}
	if len(products) == 0 && item.DefinitionID > 0 {
		products = []catalogmodel.Product{{DefinitionID: item.DefinitionID, Quantity: item.Amount}}
	}
	for _, product := range products {
		var granted []furnituremodel.Item
		var grantErr error
		extraData, extraErr := service.purchaseExtraData(ctx, params, item, product.DefinitionID)
		if extraErr != nil {
			return extraErr
		}
		grant := furnitureservice.GrantParams{DefinitionID: product.DefinitionID, OwnerPlayerID: recipientID, Quantity: product.Quantity * params.Amount, ExtraData: extraData, LimitedEditionNumber: result.LimitedUnitNumber}
		if params.Gift == nil {
			granted, grantErr = service.furniture.Grant(ctx, grant)
		} else if gifts, ok := service.furniture.(furnitureservice.GiftGranter); ok {
			granted, grantErr = gifts.GrantGift(ctx, furnitureservice.GiftGrantParams{GrantParams: grant, SpriteID: params.Gift.SpriteID, BoxID: params.Gift.BoxID, RibbonID: params.Gift.RibbonID, SenderPlayerID: params.Gift.SenderPlayerID, Message: params.Gift.Message})
		} else {
			grantErr = ErrOfferNotGiftable
		}
		if grantErr != nil {
			return fmt.Errorf("grant catalog item %d furniture: %w", item.ID, grantErr)
		}
		result.GrantedItems = append(result.GrantedItems, granted...)
	}
	if err := service.finishRewards(ctx, params.PlayerID, item, result); err != nil {
		return err
	}
	if result.LimitedUnitNumber != nil {
		if len(result.GrantedItems) == 0 {
			return ErrLimitedCompletion
		}
		completed, err := service.store.CompleteLimitedUnit(ctx, item.ID, *result.LimitedUnitNumber, params.PlayerID, result.GrantedItems[0].ID)
		if err != nil {
			return err
		}
		if !completed {
			return ErrLimitedCompletion
		}
	}
	return service.logPurchase(ctx, params, item, result, credits, points)
}

// purchaseExtraData resolves server-owned initial data for supported product layouts.
func (service *Service) purchaseExtraData(ctx context.Context, params PurchaseParams, item catalogmodel.Item, definitionID int64) (string, error) {
	definition, found := service.cache.definition(definitionID)
	if !found || definition.InteractionType != "trophy" {
		return item.ExtraData, nil
	}
	if service.players == nil || service.trophies == nil {
		return "", ErrCommerceUnavailable
	}
	buyer, found, err := service.players.FindByID(ctx, params.PlayerID)
	if err != nil {
		return "", err
	}
	if !found {
		return "", ErrInvalidPlayerID
	}
	return service.trophies.Format(buyer.Player.Username, params.ExtraData), nil
}

// DiscountedUnits returns the number of free units for a bulk amount.
func DiscountedUnits(amount int32) int32 {
	basic := amount / DiscountBatchSize
	bonus := int32(0)
	if basic >= 1 {
		if amount%DiscountBatchSize == DiscountBatchSize-1 {
			bonus = 1
		}
		bonus += basic - 1
	}
	additional := int32(0)
	for _, threshold := range AdditionalDiscountThresholds {
		if amount >= threshold {
			additional++
		}
	}
	discounted := basic + bonus + additional
	if discounted > amount {
		return amount
	}
	return discounted
}
