package pair

import (
	"context"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
)

// Furniture reads items and their definitions.
type Furniture interface {
	// FindItemByID finds one active furniture item.
	FindItemByID(context.Context, int64) (furnituremodel.Item, bool, error)
	// FindDefinitionByID finds one active furniture definition.
	FindDefinitionByID(context.Context, int64) (furnituremodel.Definition, bool, error)
}

// Store persists teleport pair relationships.
type Store interface {
	// FindByItem finds a pair containing an item.
	FindByItem(context.Context, int64) (Pair, bool, error)
	// Replace atomically removes prior relationships and stores a pair.
	Replace(context.Context, Pair) error
	// DeleteByItem removes the pair containing an item.
	DeleteByItem(context.Context, int64) (bool, error)
}

// Service validates and manages teleport relationships.
type Service struct {
	// store persists pair relationships.
	store Store
	// furniture reads item and definition records.
	furniture Furniture
}

// NewService creates teleport pair behavior.
func NewService(store Store, furniture Furniture) *Service {
	return &Service{store: store, furniture: furniture}
}

// FindTarget resolves the paired item record.
func (service *Service) FindTarget(ctx context.Context, itemID int64) (furnituremodel.Item, furnituremodel.Definition, bool, error) {
	paired, found, err := service.store.FindByItem(ctx, itemID)
	if err != nil || !found {
		return furnituremodel.Item{}, furnituremodel.Definition{}, false, err
	}
	targetID, found := paired.Other(itemID)
	if !found {
		return furnituremodel.Item{}, furnituremodel.Definition{}, false, nil
	}
	target, found, err := service.furniture.FindItemByID(ctx, targetID)
	if err != nil || !found || !target.InRoom() {
		if err == nil {
			_, err = service.store.DeleteByItem(ctx, itemID)
		}
		return furnituremodel.Item{}, furnituremodel.Definition{}, false, err
	}
	definition, found, err := service.furniture.FindDefinitionByID(ctx, target.DefinitionID)
	if err != nil || !found {
		return furnituremodel.Item{}, furnituremodel.Definition{}, false, err
	}

	return target, definition, true, nil
}

// Pair validates ownership and creates a symmetric relationship.
func (service *Service) Pair(ctx context.Context, actorPlayerID int64, firstID int64, secondID int64) (Pair, error) {
	return service.pair(ctx, actorPlayerID, firstID, secondID, true)
}

// PairGranted pairs two newly granted teleport items before room placement.
func (service *Service) PairGranted(ctx context.Context, actorPlayerID int64, firstID int64, secondID int64) (Pair, error) {
	return service.pair(ctx, actorPlayerID, firstID, secondID, false)
}

// pair validates ownership and creates one relationship with optional placement requirements.
func (service *Service) pair(ctx context.Context, actorPlayerID int64, firstID int64, secondID int64, requirePlaced bool) (Pair, error) {
	paired, err := New(firstID, secondID)
	if err != nil {
		return Pair{}, err
	}
	first, err := service.teleportItem(ctx, paired.ItemOneID, requirePlaced)
	if err != nil {
		return Pair{}, err
	}
	second, err := service.teleportItem(ctx, paired.ItemTwoID, requirePlaced)
	if err != nil {
		return Pair{}, err
	}
	if actorPlayerID > 0 && (first.OwnerPlayerID != actorPlayerID || second.OwnerPlayerID != actorPlayerID) {
		return Pair{}, ErrNotOwner
	}
	if err := service.store.Replace(ctx, paired); err != nil {
		return Pair{}, err
	}

	return paired, nil
}

// Unpair removes the relationship containing an item.
func (service *Service) Unpair(ctx context.Context, itemID int64) (bool, error) {
	if itemID <= 0 {
		return false, ErrInvalidPair
	}

	return service.store.DeleteByItem(ctx, itemID)
}

// teleportItem validates one placed teleport furniture item.
func (service *Service) teleportItem(ctx context.Context, itemID int64, requirePlaced bool) (furnituremodel.Item, error) {
	item, found, err := service.furniture.FindItemByID(ctx, itemID)
	if err != nil {
		return furnituremodel.Item{}, err
	}
	if !found || (requirePlaced && !item.InRoom()) {
		return furnituremodel.Item{}, ErrItemNotFound
	}
	definition, found, err := service.furniture.FindDefinitionByID(ctx, item.DefinitionID)
	if err != nil {
		return furnituremodel.Item{}, err
	}
	if !found || (definition.InteractionType != "teleport" && definition.InteractionType != "teleport_tile") {
		return furnituremodel.Item{}, ErrNotTeleport
	}

	return item, nil
}
