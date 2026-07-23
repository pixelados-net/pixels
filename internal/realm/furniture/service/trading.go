package service

import (
	"context"
	"errors"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	"github.com/niflaot/pixels/internal/realm/furniture/repository"
)

var (
	// ErrTradingWriterUnavailable reports missing transactional furniture support.
	ErrTradingWriterUnavailable = errors.New("furniture trading writer unavailable")
	// ErrItemUnavailable reports a failed ownership or availability guard.
	ErrItemUnavailable = errors.New("furniture item unavailable")
)

// TradingManager exposes guarded furniture operations for Marketplace and Trade.
type TradingManager interface {
	DefinitionFinder
	ItemFinder
	// ReserveForMarketplace validates and withdraws one item.
	ReserveForMarketplace(ctx context.Context, itemID int64, playerID int64) (furnituremodel.Item, furnituremodel.Definition, error)
	// ReleaseFromMarketplace restores one reserved item.
	ReleaseFromMarketplace(ctx context.Context, itemID int64, playerID int64) error
	// TransferFromMarketplace delivers one reserved item.
	TransferFromMarketplace(ctx context.Context, itemID int64, sellerID int64, buyerID int64) error
	// TransferInventoryItem transfers one unreserved item.
	TransferInventoryItem(ctx context.Context, itemID int64, fromID int64, toID int64) error
	// DeleteInventoryItem consumes one unreserved item.
	DeleteInventoryItem(ctx context.Context, itemID int64, ownerID int64) error
}

// ReserveForMarketplace validates and withdraws one item.
func (service *Service) ReserveForMarketplace(ctx context.Context, itemID int64, playerID int64) (furnituremodel.Item, furnituremodel.Definition, error) {
	item, err := service.ownedItem(ctx, itemID, playerID)
	if err != nil || !item.InInventory() || item.MarketplaceReserved || service.staged != nil && service.staged.Contains(itemID) {
		return furnituremodel.Item{}, furnituremodel.Definition{}, ErrItemUnavailable
	}
	definition, found, err := service.store.FindDefinitionByID(ctx, item.DefinitionID)
	if err != nil || !found || !definition.AllowMarketplaceSale {
		return furnituremodel.Item{}, furnituremodel.Definition{}, ErrItemUnavailable
	}
	writer, ok := service.store.(repository.TradingWriter)
	if !ok {
		return furnituremodel.Item{}, furnituremodel.Definition{}, ErrTradingWriterUnavailable
	}
	updated, err := writer.ReserveForMarketplace(ctx, itemID, playerID)
	if err != nil || !updated {
		return furnituremodel.Item{}, furnituremodel.Definition{}, ErrItemUnavailable
	}
	return item, definition, nil
}

// ReleaseFromMarketplace restores one reserved item.
func (service *Service) ReleaseFromMarketplace(ctx context.Context, itemID int64, playerID int64) error {
	return service.runTradingGuard(ctx, func(writer repository.TradingWriter) (bool, error) {
		return writer.ReleaseFromMarketplace(ctx, itemID, playerID)
	})
}

// TransferFromMarketplace delivers one reserved item.
func (service *Service) TransferFromMarketplace(ctx context.Context, itemID int64, sellerID int64, buyerID int64) error {
	return service.runTradingGuard(ctx, func(writer repository.TradingWriter) (bool, error) {
		return writer.TransferFromMarketplace(ctx, itemID, sellerID, buyerID)
	})
}

// TransferInventoryItem transfers one unreserved item.
func (service *Service) TransferInventoryItem(ctx context.Context, itemID int64, fromID int64, toID int64) error {
	return service.runTradingGuard(ctx, func(writer repository.TradingWriter) (bool, error) {
		return writer.TransferInventoryItem(ctx, itemID, fromID, toID)
	})
}

// DeleteInventoryItem consumes one unreserved item.
func (service *Service) DeleteInventoryItem(ctx context.Context, itemID int64, ownerID int64) error {
	return service.runTradingGuard(ctx, func(writer repository.TradingWriter) (bool, error) {
		return writer.DeleteInventoryItem(ctx, itemID, ownerID)
	})
}

// runTradingGuard executes one guarded trading mutation.
func (service *Service) runTradingGuard(ctx context.Context, action func(repository.TradingWriter) (bool, error)) error {
	writer, ok := service.store.(repository.TradingWriter)
	if !ok {
		return ErrTradingWriterUnavailable
	}
	updated, err := action(writer)
	if err != nil {
		return err
	}
	if !updated {
		return ErrItemUnavailable
	}
	return nil
}

var tradingManagerAssertion TradingManager = (*Service)(nil)
