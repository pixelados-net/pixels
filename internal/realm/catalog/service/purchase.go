package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"

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

// purchaseDeferralKey identifies transaction-owned post-commit collection.
type purchaseDeferralKey struct{}

// purchaseEffect stores one purchase projection deferred until commit.
type purchaseEffect struct {
	// params stores the original purchase actor.
	params PurchaseParams
	// item stores the purchased offer.
	item catalogmodel.Item
	// result stores committed grants and balances.
	result PurchaseResult
	// groupID stores the selected group reward target.
	groupID int64
	// groupProduct reports whether group projection is required.
	groupProduct bool
}

// purchaseDeferral collects effects while a broader transaction is active.
type purchaseDeferral struct {
	// effects stores ordered post-commit projections.
	effects []purchaseEffect
}

// Purchase buys one catalog offer.
func (service *Service) Purchase(ctx context.Context, params PurchaseParams) (PurchaseResult, error) {
	if params.Amount == 0 {
		params.Amount = 1
	}
	item, err := service.purchaseOffer(ctx, params)
	if err != nil {
		return PurchaseResult{}, err
	}
	if item.IsPet() {
		if service.pets == nil {
			return PurchaseResult{}, ErrCommerceUnavailable
		}
		if params.Gift != nil || params.RecipientPlayerID != 0 {
			return PurchaseResult{}, ErrOfferNotGiftable
		}
		if params.OperationKey == "" {
			params.OperationKey, err = purchaseOperationKey()
			if err != nil {
				return PurchaseResult{}, err
			}
		}
	}
	groupID, _, groupProduct, err := service.groupSelection(params, item)
	if err != nil {
		return PurchaseResult{}, err
	}
	if groupProduct && (params.Gift != nil || params.RecipientPlayerID != 0 || params.Amount != 1) {
		return PurchaseResult{}, ErrOfferNotGiftable
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

	effect := purchaseEffect{params: params, item: item, result: result, groupID: groupID, groupProduct: groupProduct}
	if deferred, ok := ctx.Value(purchaseDeferralKey{}).(*purchaseDeferral); ok {
		deferred.effects = append(deferred.effects, effect)
		return result, nil
	}
	service.projectPurchase(ctx, effect)

	return result, nil
}

// PurchaseWithin buys inside an active transaction and returns safe post-commit work.
func (service *Service) PurchaseWithin(ctx context.Context, params PurchaseParams) (PurchaseResult, func(context.Context), error) {
	deferred := &purchaseDeferral{}
	result, err := service.Purchase(context.WithValue(ctx, purchaseDeferralKey{}, deferred), params)
	if err != nil {
		return PurchaseResult{}, nil, err
	}
	complete := func(commitCtx context.Context) {
		for _, effect := range deferred.effects {
			service.projectPurchase(commitCtx, effect)
		}
	}
	return result, complete, nil
}

// PurchaseAndMutate atomically combines a purchase with a caller-owned database mutation.
func (service *Service) PurchaseAndMutate(ctx context.Context, params PurchaseParams, mutate func(context.Context, PurchaseResult) error) (PurchaseResult, error) {
	var result PurchaseResult
	var complete func(context.Context)
	err := service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		var purchaseErr error
		result, complete, purchaseErr = service.PurchaseWithin(txCtx, params)
		if purchaseErr != nil {
			return purchaseErr
		}
		return mutate(txCtx, result)
	})
	if err != nil {
		return PurchaseResult{}, err
	}
	if complete != nil {
		complete(ctx)
	}
	return result, nil
}

// projectPurchase applies cache, protocol, and event side effects after commit.
func (service *Service) projectPurchase(ctx context.Context, effect purchaseEffect) {
	params := effect.params
	item := effect.item
	result := effect.result
	service.refreshAfterLimited(ctx, item)
	if effect.groupProduct {
		itemIDs := make([]int64, len(result.GrantedItems))
		for index, granted := range result.GrantedItems {
			itemIDs[index] = granted.ID
		}
		service.groups.ProjectCatalog(ctx, params.PlayerID, effect.groupID, itemIDs)
	}
	if result.GrantedPet != nil {
		service.pets.ProjectCatalog(ctx, *result.GrantedPet)
	}
	service.publishPurchase(ctx, params.PlayerID, result)
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
	groupID, forum, groupProduct, err := service.groupSelection(params, item)
	if err != nil {
		return err
	}
	if groupProduct {
		if err = service.groups.ValidateCatalog(ctx, params.PlayerID, groupID, forum); err != nil {
			return fmt.Errorf("%w: %v", ErrGroupSelection, err)
		}
	}
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
	if item.IsService() {
		return service.logPurchase(ctx, params, item, result, credits, points)
	}
	if item.IsPet() {
		if item.PetTypeID == nil {
			return ErrOfferDisabled
		}
		reward, grantErr := service.pets.GrantCatalog(ctx, PetGrantParams{OwnerPlayerID: params.PlayerID, TypeID: *item.PetTypeID, ProductCode: item.PetProductCode, ExtraData: params.ExtraData, CatalogItemID: item.ID, OperationKey: params.OperationKey})
		if grantErr != nil {
			return fmt.Errorf("grant catalog item %d pet: %w", item.ID, grantErr)
		}
		result.GrantedPet = &reward
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
	if groupProduct {
		itemIDs := make([]int64, len(result.GrantedItems))
		for index, granted := range result.GrantedItems {
			itemIDs[index] = granted.ID
		}
		if err := service.groups.CommitCatalog(ctx, params.PlayerID, groupID, forum, itemIDs); err != nil {
			return fmt.Errorf("%w: %v", ErrGroupSelection, err)
		}
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

// groupSelection parses one server-recognized group product selection.
func (service *Service) groupSelection(params PurchaseParams, item catalogmodel.Item) (int64, bool, bool, error) {
	definition, found := service.cache.definition(item.DefinitionID)
	if !found || definition.InteractionType != "group_furniture" && definition.InteractionType != "group_forum" {
		return 0, false, false, nil
	}
	if service.groups == nil {
		return 0, false, true, ErrCommerceUnavailable
	}
	groupID, err := strconv.ParseInt(strings.TrimSpace(params.ExtraData), 10, 64)
	if err != nil || groupID <= 0 {
		return 0, false, true, ErrGroupSelection
	}
	return groupID, definition.InteractionType == "group_forum", true, nil
}

// purchaseOperationKey creates an opaque grant identity for a client purchase.
func purchaseOperationKey() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", fmt.Errorf("create catalog purchase operation key: %w", err)
	}
	return "catalog:" + hex.EncodeToString(value[:]), nil
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
