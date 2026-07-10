package admin

import (
	"context"
	"math"
	"strings"

	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
)

// CreateItem creates one catalog offer and its optional LTD stock.
func (service *Service) CreateItem(ctx context.Context, input ItemInput) (catalogmodel.Item, error) {
	item := catalogmodel.Item{PageID: input.PageID, DefinitionID: input.DefinitionID, Name: strings.TrimSpace(input.Name),
		CostCredits: input.CostCredits, CostPoints: input.CostPoints, PointsType: input.PointsType,
		Amount: input.Amount, LimitedStack: input.LimitedStack, OfferID: input.OfferID, ClubOnly: input.ClubOnly,
		OrderNum: input.OrderNum, Enabled: input.Enabled, ExtraData: input.ExtraData}
	if err := service.validateItem(ctx, item); err != nil {
		return catalogmodel.Item{}, err
	}
	var created catalogmodel.Item
	err := service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		var err error
		created, err = service.store.CreateItem(txCtx, item)
		if err != nil || created.LimitedStack == 0 {
			return err
		}

		return service.store.CreateLimitedUnits(txCtx, created.ID, created.LimitedStack)
	})
	if err != nil {
		return catalogmodel.Item{}, err
	}
	if err := service.refresh(ctx); err != nil {
		return catalogmodel.Item{}, err
	}

	return created, nil
}

// UpdateItem applies a partial catalog offer update.
func (service *Service) UpdateItem(ctx context.Context, id int64, patch ItemPatch) (catalogmodel.Item, error) {
	if id <= 0 {
		return catalogmodel.Item{}, ErrInvalidItem
	}
	item, found, err := service.store.FindItemByID(ctx, id)
	if err != nil {
		return catalogmodel.Item{}, err
	}
	if !found {
		return catalogmodel.Item{}, ErrItemNotFound
	}
	previousStack := item.LimitedStack
	applyItemPatch(&item, patch)
	if item.LimitedStack < item.LimitedSells || (item.LimitedSells > 0 && item.LimitedStack < previousStack) {
		return catalogmodel.Item{}, ErrLimitedBelowSales
	}
	if err := service.validateItem(ctx, item); err != nil {
		return catalogmodel.Item{}, err
	}
	var updated catalogmodel.Item
	err = service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		var found bool
		var err error
		updated, found, err = service.store.UpdateItem(txCtx, item)
		if err != nil {
			return err
		}
		if !found {
			return ErrConflict
		}

		return service.store.SyncLimitedUnits(txCtx, updated.ID, updated.LimitedStack)
	})
	if err != nil {
		return catalogmodel.Item{}, err
	}
	if err := service.refresh(ctx); err != nil {
		return catalogmodel.Item{}, err
	}

	return updated, nil
}

// DeleteItem soft deletes one catalog offer.
func (service *Service) DeleteItem(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrInvalidItem
	}
	item, found, err := service.store.FindItemByID(ctx, id)
	if err != nil {
		return err
	}
	if !found {
		return ErrItemNotFound
	}
	deleted, err := service.store.SoftDeleteItem(ctx, id, item.Version.Version)
	if err != nil {
		return err
	}
	if !deleted {
		return ErrConflict
	}

	return service.refresh(ctx)
}

// validateItem validates offer fields and foreign references.
func (service *Service) validateItem(ctx context.Context, item catalogmodel.Item) error {
	if item.PageID <= 0 || item.DefinitionID <= 0 || item.Name == "" || item.Amount <= 0 || item.LimitedStack < 0 ||
		item.CostCredits < 0 || item.CostPoints < 0 || item.CostCredits > math.MaxInt32 || item.CostPoints > math.MaxInt32 {
		return ErrInvalidItem
	}
	if item.PointsType == catalogmodel.CreditsType && item.CostPoints != 0 {
		return ErrInvalidItem
	}
	if item.PointsType != catalogmodel.CreditsType && item.CostCredits != 0 {
		return ErrInvalidItem
	}
	if _, found, err := service.store.FindPageByID(ctx, item.PageID); err != nil || !found {
		if err != nil {
			return err
		}
		return ErrPageNotFound
	}
	if _, found, err := service.definitions.FindDefinitionByID(ctx, item.DefinitionID); err != nil || !found {
		if err != nil {
			return err
		}
		return ErrDefinitionNotFound
	}

	return nil
}

// applyItemPatch applies present offer patch fields.
func applyItemPatch(item *catalogmodel.Item, patch ItemPatch) {
	if patch.PageID != nil {
		item.PageID = *patch.PageID
	}
	if patch.DefinitionID != nil {
		item.DefinitionID = *patch.DefinitionID
	}
	if patch.Name != nil {
		item.Name = strings.TrimSpace(*patch.Name)
	}
	if patch.CostCredits != nil {
		item.CostCredits = *patch.CostCredits
	}
	if patch.CostPoints != nil {
		item.CostPoints = *patch.CostPoints
	}
	if patch.PointsType != nil {
		item.PointsType = *patch.PointsType
	}
	if patch.Amount != nil {
		item.Amount = *patch.Amount
	}
	if patch.LimitedStack != nil {
		item.LimitedStack = *patch.LimitedStack
	}
	if patch.OfferID != nil {
		item.OfferID = *patch.OfferID
	}
	if patch.ClubOnly != nil {
		item.ClubOnly = *patch.ClubOnly
	}
	if patch.OrderNum != nil {
		item.OrderNum = *patch.OrderNum
	}
	if patch.Enabled != nil {
		item.Enabled = *patch.Enabled
	}
	if patch.ExtraData != nil {
		item.ExtraData = *patch.ExtraData
	}
}
