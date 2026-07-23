// Package tests exercises Marketplace core behavior without expanding the core file-pair limit.
package tests

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	marketcore "github.com/niflaot/pixels/internal/realm/marketplace/core"
	marketrecord "github.com/niflaot/pixels/internal/realm/marketplace/record"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
	"github.com/niflaot/pixels/pkg/redis"
)

// benchmarkStore provides deterministic search records.
type benchmarkStore struct{ searches int }

// WithinTransaction executes fixture work directly.
func (*benchmarkStore) WithinTransaction(ctx context.Context, work func(context.Context) error) error {
	return work(ctx)
}

// TokenBalance returns an empty fixture token balance.
func (*benchmarkStore) TokenBalance(context.Context, int64) (int32, error) { return 0, nil }

// AddTokens returns an empty fixture token balance.
func (*benchmarkStore) AddTokens(context.Context, int64, int32) (int32, error) { return 0, nil }

// SpendToken accepts fixture token spending.
func (*benchmarkStore) SpendToken(context.Context, int64) (bool, error) { return true, nil }

// CreateListing returns the fixture listing.
func (*benchmarkStore) CreateListing(_ context.Context, value marketrecord.Listing) (marketrecord.Listing, error) {
	return value, nil
}

// FindListingForUpdate reports no fixture listing.
func (*benchmarkStore) FindListingForUpdate(context.Context, int64) (marketrecord.Listing, bool, error) {
	return marketrecord.Listing{}, false, nil
}

// FindCheapestListing reports no fixture replacement.
func (*benchmarkStore) FindCheapestListing(context.Context, int64) (marketrecord.Listing, bool, error) {
	return marketrecord.Listing{}, false, nil
}

// MarkSold accepts a fixture sale.
func (*benchmarkStore) MarkSold(context.Context, int64, int64) (bool, error) { return true, nil }

// CloseListing reports no fixture listing.
func (*benchmarkStore) CloseListing(context.Context, int64, int64, bool) (marketrecord.Listing, bool, error) {
	return marketrecord.Listing{}, false, nil
}

// SearchOffers returns one aggregated fixture offer.
func (store *benchmarkStore) SearchOffers(context.Context, marketrecord.Search) ([]marketrecord.SearchOffer, int32, error) {
	store.searches++
	return []marketrecord.SearchOffer{{Listing: marketrecord.Listing{ID: 1, FurnitureItemID: 1, FurnitureDefinitionID: 1, RawPrice: 100, ExpiresAt: time.Now().Add(time.Hour)}, AverageRawPrice: 100, OfferCount: 2}}, 1, nil
}

// ListOwnListings returns no fixture listings.
func (*benchmarkStore) ListOwnListings(context.Context, int64, time.Time) ([]marketrecord.Listing, error) {
	return nil, nil
}

// RedeemSold returns no fixture proceeds.
func (*benchmarkStore) RedeemSold(context.Context, int64) (int64, int32, error) { return 0, 0, nil }

// ExpireListings returns no fixture expirations.
func (*benchmarkStore) ExpireListings(context.Context, int32) ([]marketrecord.Listing, error) {
	return nil, nil
}

// DefinitionStats returns no fixture history.
func (*benchmarkStore) DefinitionStats(context.Context, int64, int32) ([]marketrecord.DayStat, int32, error) {
	return nil, 0, nil
}

// benchmarkFurniture provides one definition and item.
type benchmarkFurniture struct{}

// FindDefinitionByID returns one fixture definition.
func (*benchmarkFurniture) FindDefinitionByID(context.Context, int64) (furnituremodel.Definition, bool, error) {
	return furnituremodel.Definition{SpriteID: 22, AllowMarketplaceSale: true}, true, nil
}

// ListDefinitions returns one searchable fixture definition.
func (*benchmarkFurniture) ListDefinitions(context.Context) ([]furnituremodel.Definition, error) {
	return []furnituremodel.Definition{{Base: base(1), SpriteID: 22, Name: "chair", AllowMarketplaceSale: true}}, nil
}

// FindItemByID returns one fixture furniture item.
func (*benchmarkFurniture) FindItemByID(context.Context, int64) (furnituremodel.Item, bool, error) {
	return furnituremodel.Item{Base: base(1), DefinitionID: 1, OwnerPlayerID: 1}, true, nil
}

// ListInventory returns no fixture inventory.
func (*benchmarkFurniture) ListInventory(context.Context, int64) ([]furnituremodel.Item, error) {
	return nil, nil
}

// ListRoomItems returns no fixture room furniture.
func (*benchmarkFurniture) ListRoomItems(context.Context, int64) ([]furnituremodel.Item, error) {
	return nil, nil
}

// ReserveForMarketplace returns empty fixture values.
func (*benchmarkFurniture) ReserveForMarketplace(context.Context, int64, int64) (furnituremodel.Item, furnituremodel.Definition, error) {
	return furnituremodel.Item{}, furnituremodel.Definition{}, nil
}

// ReleaseFromMarketplace accepts fixture release.
func (*benchmarkFurniture) ReleaseFromMarketplace(context.Context, int64, int64) error { return nil }

// TransferFromMarketplace accepts fixture transfer.
func (*benchmarkFurniture) TransferFromMarketplace(context.Context, int64, int64, int64) error {
	return nil
}

// TransferInventoryItem accepts fixture direct transfer.
func (*benchmarkFurniture) TransferInventoryItem(context.Context, int64, int64, int64) error {
	return nil
}

// DeleteInventoryItem accepts fixture redemption.
func (*benchmarkFurniture) DeleteInventoryItem(context.Context, int64, int64) error { return nil }

// base creates shared ids without verbose fixture fields.
func base(id int64) sharedmodel.Base { return sharedmodel.Base{Identity: sharedmodel.Identity{ID: id}} }

// service creates Marketplace behavior for benchmarks.
func service(store *benchmarkStore, cache *redis.Client) *marketcore.Service {
	return marketcore.New(marketcore.Options{Enabled: true, CommissionPercent: 1, MinimumPrice: 1, MaximumPrice: 1000, SearchCacheTTL: time.Minute}, store, &benchmarkFurniture{}, nil, cache, nil)
}

// TestBuyerPriceUsesIntegerCeiling verifies commission never truncates a partial credit.
func TestBuyerPriceUsesIntegerCeiling(t *testing.T) {
	marketplace := service(&benchmarkStore{}, nil)
	for raw, want := range map[int64]int64{1: 2, 99: 100, 100: 101, 101: 103} {
		if got := marketplace.BuyerPrice(raw); got != want {
			t.Fatalf("raw %d: got %d want %d", raw, got, want)
		}
	}
}

// TestSearchProjectsAggregates verifies average and offer-count search metadata.
func TestSearchProjectsAggregates(t *testing.T) {
	result, err := service(&benchmarkStore{}, nil).Search(context.Background(), marketcore.SearchParams{MaximumPrice: 1000})
	if err != nil || len(result.Offers) != 1 || result.Offers[0].AveragePrice != 101 || result.Offers[0].OfferCount != 2 {
		t.Fatalf("result=%#v err=%v", result, err)
	}
}

// BenchmarkSearchMiss measures a Marketplace search cache miss.
func BenchmarkSearchMiss(b *testing.B) {
	marketplace := service(&benchmarkStore{}, nil)
	b.ReportAllocs()
	for b.Loop() {
		if _, err := marketplace.Search(context.Background(), marketcore.SearchParams{MaximumPrice: 1000}); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkSearchHit measures a shared Redis Marketplace cache hit.
func BenchmarkSearchHit(b *testing.B) {
	server := miniredis.RunT(b)
	client := redis.New(redis.Config{Address: server.Addr()})
	defer client.Close()
	marketplace := service(&benchmarkStore{}, client)
	params := marketcore.SearchParams{MaximumPrice: 1000}
	if _, err := marketplace.Search(context.Background(), params); err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	b.ReportAllocs()
	for b.Loop() {
		if _, err := marketplace.Search(context.Background(), params); err != nil {
			b.Fatal(err)
		}
	}
}
