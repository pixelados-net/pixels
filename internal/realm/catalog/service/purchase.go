package service

import (
	"context"
	"errors"
	"fmt"

	catalogpurchased "github.com/niflaot/pixels/internal/realm/catalog/events/purchased"
	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/zap"
)

// Purchase buys one catalog offer.
func (service *Service) Purchase(ctx context.Context, params PurchaseParams) (PurchaseResult, error) {
	item, err := service.purchaseOffer(ctx, params)
	if err != nil {
		return PurchaseResult{}, err
	}

	result := PurchaseResult{Item: item}
	err = service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		return service.commitPurchase(txCtx, params, item, &result)
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
func (service *Service) commitPurchase(ctx context.Context, params PurchaseParams, item catalogmodel.Item, result *PurchaseResult) error {
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

	balance, err := service.charge(ctx, params.PlayerID, item)
	if err != nil {
		return err
	}
	if item.IsCredits() {
		result.NewCreditsBalance = balance
	} else {
		result.NewPointsBalance = balance
	}

	result.GrantedItems, err = service.furniture.Grant(ctx, furnitureservice.GrantParams{
		DefinitionID: item.DefinitionID, OwnerPlayerID: params.PlayerID, Quantity: item.Amount, ExtraData: item.ExtraData,
	})
	if err != nil {
		return fmt.Errorf("grant catalog item %d furniture: %w", item.ID, err)
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

	return nil
}

// charge deducts the configured offer price.
func (service *Service) charge(ctx context.Context, playerID int64, item catalogmodel.Item) (int64, error) {
	currencyType := item.PointsType
	cost := item.CostPoints
	if item.IsCredits() {
		currencyType = catalogmodel.CreditsType
		cost = item.CostCredits
	}
	if cost == 0 {
		return 0, nil
	}

	return service.currencies.Grant(ctx, currencyservice.GrantParams{
		PlayerID: playerID, CurrencyType: currencyType, Amount: -cost,
		Reason: "catalog_purchase", ActorKind: currencyservice.ActorPlayer,
	})
}

// refreshAfterLimited refreshes stock cache after an LTD purchase.
func (service *Service) refreshAfterLimited(ctx context.Context, item catalogmodel.Item) {
	if !item.IsLimited() {
		return
	}
	if err := service.Refresh(ctx); err != nil {
		service.log.Warn("catalog cache refresh after limited purchase failed", zap.Int64("catalog_item_id", item.ID), zap.Error(err))
	}
}

// publishPurchase emits a completed purchase fact.
func (service *Service) publishPurchase(ctx context.Context, playerID int64, result PurchaseResult) {
	if service.events == nil {
		return
	}
	err := service.events.Publish(ctx, bus.Event{Name: catalogpurchased.Name, Payload: catalogpurchased.Payload{
		PlayerID: playerID, CatalogItemID: result.Item.ID, DefinitionID: result.Item.DefinitionID,
		Quantity: result.Item.Amount, CostCredits: result.Item.CostCredits, CostPoints: result.Item.CostPoints,
		PointsType: result.Item.PointsType, LimitedUnitNumber: result.LimitedUnitNumber,
	}})
	if err != nil && !errors.Is(err, context.Canceled) {
		service.log.Warn("catalog purchase event projection failed", zap.Int64("player_id", playerID), zap.Int64("catalog_item_id", result.Item.ID), zap.Error(err))
	}
}
