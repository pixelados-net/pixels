package promo

import (
	"context"
	"testing"
	"time"

	progressionengine "github.com/niflaot/pixels/internal/realm/progression/engine"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
)

// promoCatalogSource returns one fixed promotional catalog.
type promoCatalogSource struct{ value progressionrecord.Catalog }

// Catalog returns the fixed promotional catalog.
func (source promoCatalogSource) Catalog(context.Context) (progressionrecord.Catalog, error) {
	return source.value, nil
}

// memoryPromoStore implements focused claim persistence.
type memoryPromoStore struct {
	progressionrecord.Store
	claims map[int64]bool
	count  int64
}

// WithinTransaction executes one in-memory claim transaction.
func (*memoryPromoStore) WithinTransaction(ctx context.Context, work func(context.Context) error) error {
	return work(ctx)
}

// ClaimPromo records one unique claim under the configured cap.
func (store *memoryPromoStore) ClaimPromo(_ context.Context, playerID int64, value progressionrecord.PromoBadge, force bool) (bool, error) {
	if store.claims[playerID] || !force && value.MaxClaims > 0 && store.count >= value.MaxClaims {
		return false, nil
	}
	store.claims[playerID] = true
	store.count++
	return true, nil
}

// PromoClaimed reports the in-memory player claim.
func (store *memoryPromoStore) PromoClaimed(_ context.Context, playerID int64, _ string) (bool, error) {
	return store.claims[playerID], nil
}

// promoBadges records awarded promotional badges.
type promoBadges struct{ count int }

// GrantBadge records one badge grant.
func (badges *promoBadges) GrantBadge(context.Context, int64, string, string) (bool, error) {
	badges.count++
	return true, nil
}

// promoFixture builds one loaded promotional service.
func promoFixture(t testing.TB, value progressionrecord.PromoBadge) (*Service, *memoryPromoStore, *promoBadges) {
	t.Helper()
	catalog := progressionengine.NewCatalog(promoCatalogSource{value: progressionrecord.Catalog{Promos: []progressionrecord.PromoBadge{value}}})
	if err := catalog.Reload(context.Background()); err != nil {
		t.Fatal(err)
	}
	store, badges := &memoryPromoStore{claims: make(map[int64]bool)}, &promoBadges{}
	return New(catalog, store, badges), store, badges
}

// TestClaimIsIdempotent verifies one player receives one promotional badge.
func TestClaimIsIdempotent(t *testing.T) {
	service, _, badges := promoFixture(t, progressionrecord.PromoBadge{Code: "PIXELS", BadgeCode: "PX", Enabled: true})
	first, err := service.Claim(context.Background(), 42, "pixels", false)
	if err != nil || !first {
		t.Fatalf("first=%v err=%v", first, err)
	}
	second, err := service.Claim(context.Background(), 42, "PIXELS", false)
	if err != nil || second || badges.count != 1 {
		t.Fatalf("second=%v grants=%d err=%v", second, badges.count, err)
	}
	claimed, err := service.Status(context.Background(), 42, "pixels")
	if err != nil || !claimed {
		t.Fatalf("claimed=%v err=%v", claimed, err)
	}
}

// TestClaimHonorsWindowAndForce verifies support overrides only availability limits.
func TestClaimHonorsWindowAndForce(t *testing.T) {
	start, end := time.Now().Add(time.Hour), time.Now().Add(2*time.Hour)
	service, _, badges := promoFixture(t, progressionrecord.PromoBadge{Code: "FUTURE", BadgeCode: "PX", StartsAt: &start, EndsAt: &end, MaxClaims: 1, Enabled: true})
	if _, err := service.Claim(context.Background(), 42, "FUTURE", false); err != progressionrecord.ErrUnavailable {
		t.Fatalf("window error %v", err)
	}
	granted, err := service.Claim(context.Background(), 42, "FUTURE", true)
	if err != nil || !granted || badges.count != 1 {
		t.Fatalf("forced=%v grants=%d err=%v", granted, badges.count, err)
	}
}

// TestClaimHonorsGlobalCap verifies the persisted cap result never grants a badge.
func TestClaimHonorsGlobalCap(t *testing.T) {
	service, store, badges := promoFixture(t, progressionrecord.PromoBadge{Code: "LIMIT", BadgeCode: "PX", MaxClaims: 1, Enabled: true})
	store.count = 1
	granted, err := service.Claim(context.Background(), 42, "LIMIT", false)
	if err != nil || granted || badges.count != 0 {
		t.Fatalf("granted=%v grants=%d err=%v", granted, badges.count, err)
	}
}
