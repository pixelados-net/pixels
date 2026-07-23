package core

import (
	"context"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/realm/subscription/record"
)

// TestCalculatePaydayUsesHighestThreshold verifies non-cumulative streak rewards.
func TestCalculatePaydayUsesHighestThreshold(t *testing.T) {
	streak, monthly := CalculatePayday(30, 250, 0.10)
	if streak != 10 || monthly != 25 {
		t.Fatalf("streak=%d monthly=%d", streak, monthly)
	}
}

// TestSubscribeExtendsFromActiveExpiration verifies prepaid time is preserved.
func TestSubscribeExtendsFromActiveExpiration(t *testing.T) {
	now := time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC)
	fixture := newCoreFixture(now)
	expires := now.Add(10 * 24 * time.Hour)
	started := now.Add(-20 * 24 * time.Hour)
	fixture.store.membership = record.Membership{PlayerID: 7, Level: record.LevelHC, StartedAt: &started, ExpiresAt: &expires}
	fixture.store.found = true
	membership, err := fixture.service.Subscribe(context.Background(), 7, record.LevelHC, 31*24*time.Hour)
	if err != nil || membership.ExpiresAt == nil || !membership.ExpiresAt.Equal(expires.Add(31*24*time.Hour)) {
		t.Fatalf("membership=%#v error=%v", membership, err)
	}
}

// TestSubscribeStartsANewTierAfterExpiration verifies stale VIP does not leak into a new HC period.
func TestSubscribeStartsANewTierAfterExpiration(t *testing.T) {
	now := time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC)
	fixture := newCoreFixture(now)
	first := now.Add(-90 * 24 * time.Hour)
	oldStreak := now.Add(-60 * 24 * time.Hour)
	expired := now.Add(-10 * 24 * time.Hour)
	fixture.store.membership = record.Membership{PlayerID: 7, Level: record.LevelVIP, StartedAt: &first,
		StreakStartedAt: &oldStreak, ExpiresAt: &expired, LastAccruedAt: &expired}
	fixture.store.found = true
	membership, err := fixture.service.Subscribe(context.Background(), 7, record.LevelHC, 31*24*time.Hour)
	if err != nil || membership.Level != record.LevelHC || membership.StartedAt == nil || !membership.StartedAt.Equal(first) ||
		membership.StreakStartedAt == nil || !membership.StreakStartedAt.Equal(now) {
		t.Fatalf("membership=%#v error=%v", membership, err)
	}
}

// TestPurchaseOfferChargesAndSubscribes verifies one transactional purchase path.
func TestPurchaseOfferChargesAndSubscribes(t *testing.T) {
	fixture := newCoreFixture(time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC))
	membership, err := fixture.service.PurchaseOffer(context.Background(), 7, 7, 1)
	if err != nil || membership.Level != record.LevelHC || len(fixture.currencies.amounts) != 1 || fixture.currencies.amounts[0] != -25 {
		t.Fatalf("membership=%#v amounts=%#v error=%v", membership, fixture.currencies.amounts, err)
	}
}

// TestPurchaseOfferAmountMultipliesPriceAndDuration verifies catalog quantities.
func TestPurchaseOfferAmountMultipliesPriceAndDuration(t *testing.T) {
	now := time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC)
	fixture := newCoreFixture(now)
	membership, err := fixture.service.PurchaseOfferAmount(context.Background(), 7, 7, 1, 2)
	if err != nil || membership.ExpiresAt == nil || !membership.ExpiresAt.Equal(now.Add(62*24*time.Hour)) ||
		len(fixture.currencies.amounts) != 1 || fixture.currencies.amounts[0] != -50 {
		t.Fatalf("membership=%#v amounts=%#v error=%v", membership, fixture.currencies.amounts, err)
	}
	if _, err := fixture.service.PurchaseOfferAmount(context.Background(), 7, 7, 1, 0); err != ErrInvalidAmount {
		t.Fatalf("expected invalid amount, got %v", err)
	}
}

// TestPurchaseExtensionOfferEnforcesDealTier verifies dedicated extension packets cannot select arbitrary offers.
func TestPurchaseExtensionOfferEnforcesDealTier(t *testing.T) {
	fixture := newCoreFixture(time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC))
	fixture.store.offers = append(fixture.store.offers, record.Offer{ID: 2, Name: "hc_90_days",
		DayCount: 90, PriceCredits: 65, Deal: true, Enabled: true})
	if _, err := fixture.service.PurchaseExtensionOffer(context.Background(), 7, 2, false); err != nil {
		t.Fatalf("purchase extension: %v", err)
	}
	if _, err := fixture.service.PurchaseExtensionOffer(context.Background(), 7, 2, true); err != ErrOfferNotFound {
		t.Fatalf("expected tier rejection, got %v", err)
	}
	if _, err := fixture.service.PurchaseExtensionOffer(context.Background(), 7, 1, false); err != ErrOfferNotFound {
		t.Fatalf("expected non-deal rejection, got %v", err)
	}
}

// TestRemainingClubGiftsSubtractsClaims verifies monthly gift idempotency accounting.
func TestRemainingClubGiftsSubtractsClaims(t *testing.T) {
	membership := record.Membership{LifetimeActiveSeconds: 2 * ClubGiftPeriodSeconds, GiftsClaimed: 1}
	if remaining := RemainingClubGifts(membership); remaining != 1 {
		t.Fatalf("remaining=%d", remaining)
	}
}

// BenchmarkCalculatePayday measures pure payday calculations.
func BenchmarkCalculatePayday(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		streak, monthly := CalculatePayday(365, 10_000, 0.1)
		if streak != 30 || monthly != 1_000 {
			b.Fatal("unexpected payday")
		}
	}
}

// TestDiscountBoundary verifies monthly gift period boundaries.
func TestDiscountBoundary(t *testing.T) {
	if remaining := RemainingClubGifts(record.Membership{LifetimeActiveSeconds: ClubGiftPeriodSeconds - 1}); remaining != 0 {
		t.Fatalf("remaining=%d", remaining)
	}
}

// TestCurrentPaydayInfoUsesNitroMinutes verifies the protocol countdown unit.
func TestCurrentPaydayInfoUsesNitroMinutes(t *testing.T) {
	now := time.Date(2026, 7, 12, 12, 0, 30, 0, time.UTC)
	fixture := newCoreFixture(now)
	started := now.Add(-8 * 24 * time.Hour)
	lastPayday := now.Add(-30 * 24 * time.Hour)
	fixture.store.membership = record.Membership{PlayerID: 7, Level: record.LevelHC, StartedAt: &started,
		StreakStartedAt: &started, LastPaydayAt: &lastPayday}
	fixture.store.found = true
	info, err := fixture.service.CurrentPaydayInfo(context.Background(), 7)
	if err != nil || info.MinutesUntilPayday != 1440 || info.StreakDays != 8 {
		t.Fatalf("info=%#v error=%v", info, err)
	}
}

// TestRunCycleBackfillsDowntimeAndEveryMissedPayday verifies durable catch-up.
func TestRunCycleBackfillsDowntimeAndEveryMissedPayday(t *testing.T) {
	now := time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC)
	fixture := newCoreFixture(now)
	started := now.Add(-70 * 24 * time.Hour)
	lastPayday := now.Add(-63 * 24 * time.Hour)
	lastAccrued := now.Add(-5 * 24 * time.Hour)
	expires := now.Add(24 * time.Hour)
	fixture.store.membership = record.Membership{PlayerID: 7, Level: record.LevelHC, StartedAt: &started,
		StreakStartedAt: &started, ExpiresAt: &expires, LastPaydayAt: &lastPayday,
		LastAccruedAt: &lastAccrued, LifetimeActiveSeconds: 58 * 86_400}
	fixture.store.found = true
	fixture.store.active = []record.Membership{fixture.store.membership}
	if err := fixture.service.RunCycle(context.Background()); err != nil {
		t.Fatalf("run catch-up cycle: %v", err)
	}
	if len(fixture.store.paydays) != 2 || fixture.store.membership.LifetimeActiveSeconds != 63*86_400 {
		t.Fatalf("membership=%#v paydays=%#v", fixture.store.membership, fixture.store.paydays)
	}
}

// TestMembershipProjectsAccruedTimeWithoutWriteAmplification verifies read-time freshness.
func TestMembershipProjectsAccruedTimeWithoutWriteAmplification(t *testing.T) {
	now := time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC)
	fixture := newCoreFixture(now)
	started := now.Add(-24 * time.Hour)
	expires := now.Add(24 * time.Hour)
	fixture.store.membership = record.Membership{PlayerID: 7, Level: record.LevelVIP, StartedAt: &started,
		StreakStartedAt: &started, LastAccruedAt: &started, ExpiresAt: &expires}
	fixture.store.found = true
	projected, found, err := fixture.service.Membership(context.Background(), 7)
	if err != nil || !found || projected.LifetimeActiveSeconds != 86_400 || projected.LifetimeVIPSeconds != 86_400 {
		t.Fatalf("projected=%#v found=%t error=%v", projected, found, err)
	}
	if fixture.store.membership.LifetimeActiveSeconds != 0 || fixture.store.membership.Version != 0 {
		t.Fatalf("read unexpectedly persisted membership=%#v", fixture.store.membership)
	}
}
