package core

import (
	"context"
	"encoding/json"
	"math"
	"strings"

	marketrecord "github.com/niflaot/pixels/internal/realm/marketplace/record"
)

// Search returns at most 250 commission-filtered Marketplace offers.
func (service *Service) Search(ctx context.Context, params SearchParams) (SearchResult, error) {
	if !service.config.Enabled {
		return SearchResult{}, ErrDisabled
	}
	if params.MinimumPrice < 0 {
		params.MinimumPrice = 0
	}
	if params.MaximumPrice <= 0 {
		params.MaximumPrice = math.MaxInt64
	}
	key := service.versionedCacheKey(ctx, params)
	if service.cache != nil {
		if encoded, found, err := service.cache.Find(ctx, key); err == nil && found {
			var cached SearchResult
			if json.Unmarshal(encoded, &cached) == nil {
				return cached, nil
			}
		}
	}
	definitions, err := service.furniture.ListDefinitions(ctx)
	if err != nil {
		return SearchResult{}, err
	}
	query := strings.ToLower(strings.TrimSpace(params.Query))
	definitionIDs := make([]int64, 0, len(definitions))
	for _, definition := range definitions {
		if query == "" || strings.Contains(strings.ToLower(definition.Name), query) || strings.Contains(strings.ToLower(definition.PublicName), query) {
			definitionIDs = append(definitionIDs, definition.ID)
		}
	}
	if query != "" && len(definitionIDs) == 0 {
		return SearchResult{}, nil
	}
	searchMinimum := params.MinimumPrice
	if searchMinimum < service.BuyerPrice(service.config.MinimumPrice) {
		searchMinimum = service.BuyerPrice(service.config.MinimumPrice)
	}
	searchMaximum := params.MaximumPrice
	if configuredMaximum := service.BuyerPrice(service.config.MaximumPrice); searchMaximum > configuredMaximum {
		searchMaximum = configuredMaximum
	}
	groups, total, err := service.store.SearchOffers(ctx, marketrecord.Search{MinimumBuyerPrice: searchMinimum, MaximumBuyerPrice: searchMaximum, CommissionPercent: service.config.CommissionPercent, DefinitionIDs: definitionIDs, SortType: params.SortType, Limit: 250})
	if err != nil {
		return SearchResult{}, err
	}
	result := SearchResult{Offers: make([]Offer, 0, len(groups)), Total: total}
	definitionsIndex := make(map[int64]int, len(definitions))
	for index := range definitions {
		definitionsIndex[definitions[index].ID] = index
	}
	for _, group := range groups {
		listing := group.Listing
		price := service.BuyerPrice(listing.RawPrice)
		item, found, findErr := service.furniture.FindItemByID(ctx, listing.FurnitureItemID)
		if findErr != nil {
			return SearchResult{}, findErr
		}
		if !found {
			continue
		}
		index, found := definitionsIndex[listing.FurnitureDefinitionID]
		if !found {
			continue
		}
		result.Offers = append(result.Offers, Offer{Listing: listing, Item: item, Definition: definitions[index], BuyerPrice: price, MinutesRemaining: nowMinutes(listing.ExpiresAt, service.now()), AveragePrice: service.BuyerPrice(group.AverageRawPrice), OfferCount: group.OfferCount})
	}
	if service.cache != nil {
		if encoded, marshalErr := json.Marshal(result); marshalErr == nil {
			_ = service.cache.Set(ctx, key, encoded, service.config.SearchCacheTTL)
		}
	}
	return result, nil
}

// OwnListings returns the seller's current and historical listing projection.
func (service *Service) OwnListings(ctx context.Context, sellerID int64) ([]Offer, int64, error) {
	listings, err := service.store.ListOwnListings(ctx, sellerID, service.now().Add(-service.config.DisplayDuration))
	if err != nil {
		return nil, 0, err
	}
	offers := make([]Offer, 0, len(listings))
	var waiting int64
	definitions, err := service.furniture.ListDefinitions(ctx)
	if err != nil {
		return nil, 0, err
	}
	definitionsIndex := make(map[int64]int, len(definitions))
	for index := range definitions {
		definitionsIndex[definitions[index].ID] = index
	}
	for _, listing := range listings {
		if listing.State == marketrecord.StateSold && listing.RedeemedAt == nil {
			waiting += listing.RawPrice
		}
		item, found, findErr := service.furniture.FindItemByID(ctx, listing.FurnitureItemID)
		if findErr != nil {
			return nil, 0, findErr
		}
		index, defined := definitionsIndex[listing.FurnitureDefinitionID]
		if !found || !defined {
			continue
		}
		offers = append(offers, Offer{Listing: listing, Item: item, Definition: definitions[index], BuyerPrice: service.BuyerPrice(listing.RawPrice), MinutesRemaining: nowMinutes(listing.ExpiresAt, service.now())})
	}
	return offers, waiting, nil
}

// ItemStats returns recent statistics for a furniture definition.
func (service *Service) ItemStats(ctx context.Context, definitionID int64) (Stats, error) {
	history, count, err := service.store.DefinitionStats(ctx, definitionID, 30)
	if err != nil {
		return Stats{}, err
	}
	var average int64
	if len(history) > 0 {
		average = service.BuyerPrice(history[0].AverageRawPrice)
	}
	for index := range history {
		history[index].AverageRawPrice = service.BuyerPrice(history[index].AverageRawPrice)
	}
	return Stats{AveragePrice: average, OpenCount: count, History: history}, nil
}

// DefinitionIDBySprite resolves a furniture definition without requiring an active listing.
func (service *Service) DefinitionIDBySprite(ctx context.Context, spriteID int32) (int64, bool, error) {
	definitions, err := service.furniture.ListDefinitions(ctx)
	if err != nil {
		return 0, false, err
	}
	for _, definition := range definitions {
		if definition.SpriteID == int(spriteID) {
			return definition.ID, true, nil
		}
	}
	return 0, false, nil
}

// versionedCacheKey includes the current shared invalidation generation.
func (service *Service) versionedCacheKey(ctx context.Context, params SearchParams) string {
	generation := "0"
	if service.cache != nil {
		if value, found, err := service.cache.Find(ctx, "marketplace:search:generation"); err == nil && found {
			generation = string(value)
		}
	}
	return service.cacheKey(params) + ":" + generation
}
