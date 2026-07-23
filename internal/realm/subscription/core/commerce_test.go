package core

import (
	"context"
	"errors"
	"testing"
	"time"

	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	"github.com/niflaot/pixels/internal/realm/subscription/record"
)

// TestTargetedOfferLifecycle verifies selection, state, limit, and purchase behavior.
func TestTargetedOfferLifecycle(t *testing.T) {
	fixture := newCoreFixture(time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC))
	fixture.store.targeted = record.TargetedOffer{ID: 8, CatalogItemID: 3, PurchaseLimit: 1, Enabled: true}
	offer, found, err := fixture.service.TargetedOffer(context.Background(), 7, 0)
	if err != nil || !found || offer.ID != 8 {
		t.Fatalf("offer=%#v found=%t error=%v", offer, found, err)
	}
	if _, err := fixture.service.PurchaseTargetedOffer(context.Background(), 7, 8, 1); err != nil {
		t.Fatalf("purchase targeted offer: %v", err)
	}
	if _, err := fixture.service.PurchaseTargetedOffer(context.Background(), 7, 8, 1); !errors.Is(err, ErrTargetedOfferUnavailable) {
		t.Fatalf("expected unavailable offer, got %v", err)
	}
	if err := fixture.service.SetTargetedState(context.Background(), 7, 8, true); err != nil {
		t.Fatalf("dismiss targeted offer: %v", err)
	}
	if _, found, err := fixture.service.TargetedOffer(context.Background(), 7, 0); err != nil || found {
		t.Fatalf("dismissed offer found=%t error=%v", found, err)
	}
}

// TestTargetedOfferQuantityConsumesTheWholeLimit verifies atomic quantity accounting.
func TestTargetedOfferQuantityConsumesTheWholeLimit(t *testing.T) {
	fixture := newCoreFixture(time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC))
	fixture.store.targeted = record.TargetedOffer{ID: 8, CatalogItemID: 3, PurchaseLimit: 3, Enabled: true}
	if _, err := fixture.service.PurchaseTargetedOffer(context.Background(), 7, 8, 2); err != nil {
		t.Fatalf("purchase targeted quantity: %v", err)
	}
	if fixture.store.targetedPurchases != 2 {
		t.Fatalf("purchases=%d", fixture.store.targetedPurchases)
	}
	if _, err := fixture.service.PurchaseTargetedOffer(context.Background(), 7, 8, 2); !errors.Is(err, ErrTargetedOfferUnavailable) {
		t.Fatalf("expected quantity limit rejection, got %v", err)
	}
}

// TestClaimClubGiftRequiresActiveMembershipAndGiftPage verifies service-level claim gates.
func TestClaimClubGiftRequiresActiveMembershipAndGiftPage(t *testing.T) {
	now := time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC)
	fixture := newCoreFixture(now)
	started := now.Add(-32 * 24 * time.Hour)
	expires := now.Add(24 * time.Hour)
	fixture.store.membership = record.Membership{PlayerID: 7, Level: record.LevelHC, StartedAt: &started,
		ExpiresAt: &expires, LifetimeActiveSeconds: ClubGiftPeriodSeconds}
	fixture.store.found = true
	page := catalogmodel.Page{Layout: "club_gifts"}
	page.ID = 102
	item := catalogmodel.Item{PageID: 102, ExtraData: "0"}
	item.ID = 1003
	fixture.catalog.pages = []catalogmodel.Page{page}
	fixture.catalog.items = map[int64][]catalogmodel.Item{102: {item}}
	if _, err := fixture.service.ClaimClubGift(context.Background(), 7, 1003); err != nil {
		t.Fatalf("claim valid club gift: %v", err)
	}
	if fixture.store.membership.GiftsClaimed != 1 || !fixture.store.giftClaimed {
		t.Fatalf("membership=%#v claimed=%t", fixture.store.membership, fixture.store.giftClaimed)
	}
	fixture.store.giftClaimed = false
	fixture.store.membership.GiftsClaimed = 0
	fixture.store.membership.ExpiresAt = &started
	if _, err := fixture.service.ClaimClubGift(context.Background(), 7, 1003); !errors.Is(err, ErrMembershipNotFound) {
		t.Fatalf("expected expired membership rejection, got %v", err)
	}
	fixture.store.membership.ExpiresAt = &expires
	if _, err := fixture.service.ClaimClubGift(context.Background(), 7, 3); !errors.Is(err, ErrOfferNotFound) {
		t.Fatalf("expected non-gift offer rejection, got %v", err)
	}
}

// TestCalendarDoorAppliesDateAndClaimGuards verifies normal and staff calendar access.
func TestCalendarDoorAppliesDateAndClaimGuards(t *testing.T) {
	now := time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC)
	fixture := newCoreFixture(now)
	fixture.store.campaign = record.Campaign{ID: 4, Name: "summer", StartDate: now, DayCount: 3, Enabled: true}
	fixture.store.campaignDay = record.CampaignDay{CampaignID: 4, DayNumber: 1, CreditsReward: 5}
	if _, err := fixture.service.OpenCalendarDoor(context.Background(), 7, "summer", 1, false); !errors.Is(err, ErrCalendarDoorUnavailable) {
		t.Fatalf("expected future door rejection, got %v", err)
	}
	if _, err := fixture.service.OpenCalendarDoor(context.Background(), 7, "summer", 1, true); err != nil {
		t.Fatalf("staff open calendar door: %v", err)
	}
	if len(fixture.currencies.amounts) != 1 || fixture.currencies.amounts[0] != 5 {
		t.Fatalf("unexpected calendar rewards %#v", fixture.currencies.amounts)
	}
	if _, err := fixture.service.OpenCalendarDoor(context.Background(), 7, "summer", 1, true); !errors.Is(err, ErrCalendarDoorUnavailable) {
		t.Fatalf("expected duplicate door rejection, got %v", err)
	}
	campaign, days, opened, err := fixture.service.CalendarData(context.Background(), 7, "summer")
	if err != nil || campaign.ID != 4 || len(days) != 1 || len(opened) != 1 {
		t.Fatalf("campaign=%#v days=%#v opened=%#v error=%v", campaign, days, opened, err)
	}
}

// TestSchedulerRecordsAndClaimsPayday verifies accrual and exactly-once claim state.
func TestSchedulerRecordsAndClaimsPayday(t *testing.T) {
	now := time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC)
	fixture := newCoreFixture(now)
	started := now.Add(-40 * 24 * time.Hour)
	streak := started
	lastPayday := now.Add(-32 * 24 * time.Hour)
	lastAccrued := now.Add(-time.Hour)
	expires := now.Add(24 * time.Hour)
	fixture.store.membership = record.Membership{PlayerID: 7, Level: record.LevelHC, StartedAt: &started,
		StreakStartedAt: &streak, ExpiresAt: &expires, LastPaydayAt: &lastPayday, LastAccruedAt: &lastAccrued}
	fixture.store.found = true
	fixture.store.active = []record.Membership{fixture.store.membership}
	if err := fixture.service.RunCycle(context.Background()); err != nil {
		t.Fatalf("run subscription cycle: %v", err)
	}
	if len(fixture.store.paydays) != 1 || fixture.store.paydays[0].TotalAwarded != 20 {
		t.Fatalf("unexpected paydays %#v", fixture.store.paydays)
	}
	if err := fixture.service.ClaimPaydays(context.Background(), 7); err != nil {
		t.Fatalf("claim paydays: %v", err)
	}
	if !fixture.store.paydays[0].Claimed || fixture.currencies.amounts[len(fixture.currencies.amounts)-1] != 20 {
		t.Fatalf("payday not claimed: %#v rewards=%#v", fixture.store.paydays, fixture.currencies.amounts)
	}
}

// TestRunCycleExpiresMembership verifies expired club projection removal.
func TestRunCycleExpiresMembership(t *testing.T) {
	now := time.Date(2026, 7, 12, 12, 0, 0, 0, time.UTC)
	fixture := newCoreFixture(now)
	expires := now.Add(-time.Second)
	fixture.store.membership = record.Membership{PlayerID: 7, Level: record.LevelVIP, ExpiresAt: &expires}
	fixture.store.found = true
	fixture.store.active = []record.Membership{fixture.store.membership}
	if err := fixture.service.RunCycle(context.Background()); err != nil {
		t.Fatalf("expire membership: %v", err)
	}
	if fixture.store.membership.Level != record.LevelNone || fixture.players.club.Level != 0 {
		t.Fatalf("membership=%#v club=%#v", fixture.store.membership, fixture.players.club)
	}
}
