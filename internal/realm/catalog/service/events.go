package service

import (
	"context"
	"errors"
	"fmt"

	catalogpurchased "github.com/niflaot/pixels/internal/realm/catalog/events/purchased"
	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	roompurchased "github.com/niflaot/pixels/internal/realm/room/record/events/bundlepurchased"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/zap"
)

// logPurchase writes extended commerce history when enabled.
func (service *Service) logPurchase(ctx context.Context, params PurchaseParams, item catalogmodel.Item, result *PurchaseResult, credits int64, points int64) error {
	if params.Free || service.commerce == nil {
		return nil
	}
	itemIDs := make([]int64, len(result.GrantedItems))
	for index, granted := range result.GrantedItems {
		itemIDs[index] = granted.ID
	}
	if err := service.commerce.LogPurchase(ctx, params.PlayerID, item, params.Amount, credits, points, itemIDs); err != nil {
		return fmt.Errorf("log catalog purchase: %w", err)
	}
	return nil
}

// publishPurchase emits completed catalog and room bundle purchase facts.
func (service *Service) publishPurchase(ctx context.Context, playerID int64, result PurchaseResult) {
	if service.events == nil {
		return
	}
	err := service.events.Publish(ctx, bus.Event{Name: catalogpurchased.Name, Payload: catalogpurchased.Payload{
		PlayerID: playerID, CatalogItemID: result.Item.ID, DefinitionID: result.Item.DefinitionID,
		Quantity: result.Item.Amount, CostCredits: result.Item.CostCredits, CostPoints: result.Item.CostPoints,
		PointsType: result.Item.PointsType, LimitedUnitNumber: result.LimitedUnitNumber, CreatedRoomID: result.CreatedRoomID,
	}})
	if err != nil && !errors.Is(err, context.Canceled) {
		service.log.Warn("catalog purchase event projection failed", zap.Int64("player_id", playerID), zap.Int64("catalog_item_id", result.Item.ID), zap.Error(err))
	}
	if result.CreatedRoomID == nil || result.Item.RoomBundleTemplateRoomID == nil {
		return
	}
	err = service.events.Publish(ctx, bus.Event{Name: roompurchased.Name, Payload: roompurchased.Payload{
		PlayerID: playerID, CatalogItemID: result.Item.ID, TemplateRoomID: *result.Item.RoomBundleTemplateRoomID,
		CreatedRoomID: *result.CreatedRoomID, FurnitureCount: result.ClonedFurnitureCount, BotCount: result.ClonedBotCount,
	}})
	if err != nil && !errors.Is(err, context.Canceled) {
		service.log.Warn("room bundle purchase event projection failed", zap.Int64("player_id", playerID), zap.Error(err))
	}
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
