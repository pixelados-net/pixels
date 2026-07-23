package core

import (
	"context"
	"fmt"
	"strings"
	"time"

	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	marketlisted "github.com/niflaot/pixels/internal/realm/marketplace/events/listed"
	marketsold "github.com/niflaot/pixels/internal/realm/marketplace/events/sold"
	marketrecord "github.com/niflaot/pixels/internal/realm/marketplace/record"
	"github.com/niflaot/pixels/pkg/bus"
	"github.com/niflaot/pixels/pkg/redis"
)

const creditsType int32 = -1

// Service implements Marketplace business behavior.
type Service struct {
	// config stores immutable Marketplace policy.
	config Options
	// store persists listings, tokens, and statistics.
	store marketrecord.Store
	// furniture coordinates ownership and Marketplace limbo.
	furniture furnitureservice.TradingManager
	// currencies mutates credit balances.
	currencies currencyservice.Granter
	// cache stores shared short-lived search results.
	cache *redis.Client
	// now supplies current time for deterministic tests.
	now func() time.Time
	// events publishes committed Marketplace facts.
	events bus.Publisher
}

// New creates a Marketplace service.
func New(config Options, store marketrecord.Store, furniture furnitureservice.TradingManager, currencies currencyservice.Granter, cache *redis.Client, events bus.Publisher) *Service {
	return &Service{config: config, store: store, furniture: furniture, currencies: currencies, cache: cache, now: time.Now, events: events}
}

// BuyerPrice adds commission with integer ceiling semantics.
func (service *Service) BuyerPrice(raw int64) int64 {
	return raw + (raw*service.config.CommissionPercent+99)/100
}

// Config returns immutable Marketplace configuration.
func (service *Service) Config() Options { return service.config }

// CanSell reports whether Marketplace is enabled.
func (service *Service) CanSell() bool { return service.config.Enabled }

// Tokens returns a player's current listing-token balance.
func (service *Service) Tokens(ctx context.Context, playerID int64) (int32, error) {
	return service.store.TokenBalance(ctx, playerID)
}

// BuyTokens purchases one configured token package with credits.
func (service *Service) BuyTokens(ctx context.Context, playerID int64) (int32, error) {
	if !service.config.Enabled {
		return 0, ErrDisabled
	}
	var balance int32
	err := service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		if _, err := service.currencies.Grant(txCtx, currencyservice.GrantParams{PlayerID: playerID, CurrencyType: creditsType, Amount: -service.config.TokenCost, Reason: "marketplace_tokens", ActorKind: currencyservice.ActorPlayer}); err != nil {
			return err
		}
		var err error
		balance, err = service.store.AddTokens(txCtx, playerID, service.config.TokenPackageSize)
		return err
	})
	return balance, err
}

// List creates a listing and withdraws its furniture atomically.
func (service *Service) List(ctx context.Context, playerID int64, itemID int64, rawPrice int64) (marketrecord.Listing, error) {
	if !service.config.Enabled {
		return marketrecord.Listing{}, ErrDisabled
	}
	if rawPrice < service.config.MinimumPrice || rawPrice > service.config.MaximumPrice {
		return marketrecord.Listing{}, ErrInvalidPrice
	}
	var listing marketrecord.Listing
	err := service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		spent, err := service.store.SpendToken(txCtx, playerID)
		if err != nil {
			return err
		}
		if !spent {
			return ErrNoToken
		}
		item, definition, err := service.furniture.ReserveForMarketplace(txCtx, itemID, playerID)
		if err != nil {
			return err
		}
		listing, err = service.store.CreateListing(txCtx, marketrecord.Listing{SellerPlayerID: playerID, FurnitureItemID: item.ID, FurnitureDefinitionID: definition.ID, RawPrice: rawPrice, ExpiresAt: service.now().Add(service.config.OfferDuration)})
		return err
	})
	service.invalidate(ctx)
	if err == nil && service.events != nil {
		_ = service.events.Publish(ctx, bus.Event{Name: marketlisted.Name, Payload: marketlisted.Payload{ListingID: listing.ID, SellerPlayerID: playerID, FurnitureItemID: itemID, FurnitureDefinitionID: listing.FurnitureDefinitionID, RawPrice: rawPrice, ExpiresAt: listing.ExpiresAt}})
	}
	return listing, err
}

// Buy purchases one open listing exactly once.
func (service *Service) Buy(ctx context.Context, buyerID int64, listingID int64) (marketrecord.Listing, error) {
	if !service.config.Enabled {
		return marketrecord.Listing{}, ErrDisabled
	}
	var listing marketrecord.Listing
	err := service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		var found bool
		var err error
		listing, found, err = service.store.FindListingForUpdate(txCtx, listingID)
		if err != nil {
			return err
		}
		if !found || listing.State != marketrecord.StateOpen || !listing.ExpiresAt.After(service.now()) {
			return ErrListingUnavailable
		}
		if listing.SellerPlayerID == buyerID {
			return ErrOwnListing
		}
		if _, err = service.currencies.Grant(txCtx, currencyservice.GrantParams{PlayerID: buyerID, CurrencyType: creditsType, Amount: -service.BuyerPrice(listing.RawPrice), Reason: "marketplace_purchase", ActorKind: currencyservice.ActorPlayer}); err != nil {
			return err
		}
		if err = service.furniture.TransferFromMarketplace(txCtx, listing.FurnitureItemID, listing.SellerPlayerID, buyerID); err != nil {
			return err
		}
		updated, err := service.store.MarkSold(txCtx, listing.ID, buyerID)
		if err != nil {
			return err
		}
		if !updated {
			return ErrListingUnavailable
		}
		return nil
	})
	service.invalidate(ctx)
	if err == nil && service.events != nil {
		_ = service.events.Publish(ctx, bus.Event{Name: marketsold.Name, Payload: marketsold.Payload{ListingID: listing.ID, SellerPlayerID: listing.SellerPlayerID, BuyerPlayerID: buyerID, FurnitureItemID: listing.FurnitureItemID, RawPrice: listing.RawPrice, BuyerPrice: service.BuyerPrice(listing.RawPrice)}})
	}
	return listing, err
}

// Replacement returns the cheapest current listing for an unavailable offer's definition.
func (service *Service) Replacement(ctx context.Context, unavailable marketrecord.Listing) (marketrecord.Listing, bool, error) {
	if unavailable.FurnitureDefinitionID <= 0 {
		return marketrecord.Listing{}, false, nil
	}
	return service.store.FindCheapestListing(ctx, unavailable.FurnitureDefinitionID)
}

// Close cancels one seller listing or force-closes it administratively.
func (service *Service) Close(ctx context.Context, listingID int64, sellerID int64, force bool) error {
	err := service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		listing, closed, err := service.store.CloseListing(txCtx, listingID, sellerID, force)
		if err != nil {
			return err
		}
		if !closed {
			return ErrListingUnavailable
		}
		return service.furniture.ReleaseFromMarketplace(txCtx, listing.FurnitureItemID, listing.SellerPlayerID)
	})
	service.invalidate(ctx)
	return err
}

// Redeem grants raw proceeds for every newly redeemed sold listing.
func (service *Service) Redeem(ctx context.Context, sellerID int64) (int64, int32, error) {
	var total int64
	var count int32
	err := service.store.WithinTransaction(ctx, func(txCtx context.Context) error {
		var err error
		total, count, err = service.store.RedeemSold(txCtx, sellerID)
		if err != nil || total == 0 {
			return err
		}
		_, err = service.currencies.Grant(txCtx, currencyservice.GrantParams{PlayerID: sellerID, CurrencyType: creditsType, Amount: total, Reason: "marketplace_redeem", ActorKind: currencyservice.ActorPlayer})
		return err
	})
	return total, count, err
}

// cacheKey returns a normalized search cache key.
func (service *Service) cacheKey(params SearchParams) string {
	return fmt.Sprintf("marketplace:search:%d:%d:%d:%s", params.MinimumPrice, params.MaximumPrice, params.SortType, strings.ToLower(strings.TrimSpace(params.Query)))
}

// invalidate advances the shared cache generation.
func (service *Service) invalidate(ctx context.Context) {
	if service.cache != nil {
		_, _ = service.cache.Increment(ctx, "marketplace:search:generation", 24*time.Hour)
	}
}
