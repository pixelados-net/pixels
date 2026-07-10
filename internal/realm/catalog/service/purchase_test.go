package service

import (
	"context"
	"errors"
	"sync"
	"testing"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
)

// TestPurchaseCreditsChargesAndGrants verifies a regular credits purchase.
func TestPurchaseCreditsChargesAndGrants(t *testing.T) {
	fixture := newServiceFixture(t, itemForTest())
	result, err := fixture.service.Purchase(context.Background(), PurchaseParams{PlayerID: 7, CatalogItemID: 10, Rank: 1})
	if err != nil {
		t.Fatalf("purchase credits: %v", err)
	}
	if result.NewCreditsBalance != 90 || len(result.GrantedItems) != 1 {
		t.Fatalf("unexpected purchase result %#v", result)
	}
	if len(fixture.currency.calls) != 1 || fixture.currency.calls[0].CurrencyType != -1 || fixture.currency.calls[0].Amount != -10 {
		t.Fatalf("unexpected currency calls %#v", fixture.currency.calls)
	}
	if len(fixture.furniture.calls) != 1 || fixture.furniture.calls[0].DefinitionID != 2 {
		t.Fatalf("unexpected furniture calls %#v", fixture.furniture.calls)
	}
}

// TestPurchasePointsChargesConfiguredCurrency verifies an activity-points purchase.
func TestPurchasePointsChargesConfiguredCurrency(t *testing.T) {
	item := itemForTest()
	item.PointsType = 5
	item.CostCredits = 0
	item.CostPoints = 4
	fixture := newServiceFixture(t, item)

	result, err := fixture.service.Purchase(context.Background(), PurchaseParams{PlayerID: 7, CatalogItemID: 10, Rank: 1})
	if err != nil || result.NewPointsBalance != 96 {
		t.Fatalf("unexpected result=%#v err=%v", result, err)
	}
	if fixture.currency.calls[0].CurrencyType != 5 || fixture.currency.calls[0].Amount != -4 {
		t.Fatalf("unexpected points charge %#v", fixture.currency.calls[0])
	}
}

// TestPurchaseInsufficientBalanceRollsBackReservation verifies failed charges grant nothing.
func TestPurchaseInsufficientBalanceRollsBackReservation(t *testing.T) {
	item := itemForTest()
	item.LimitedStack = 1
	fixture := newServiceFixture(t, item)
	fixture.currency.err = currencyservice.ErrInsufficientBalance

	_, err := fixture.service.Purchase(context.Background(), PurchaseParams{PlayerID: 7, CatalogItemID: 10, Rank: 1})
	if !errors.Is(err, currencyservice.ErrInsufficientBalance) {
		t.Fatalf("expected insufficient balance, got %v", err)
	}
	if fixture.store.reserved || len(fixture.furniture.calls) != 0 || len(fixture.currency.calls) != 0 {
		t.Fatalf("expected rollback store=%#v grants=%#v", fixture.store, fixture.furniture.calls)
	}
}

// TestPurchaseLimitedCompletesEditionAndRejectsSoldOut verifies LTD lifecycle.
func TestPurchaseLimitedCompletesEditionAndRejectsSoldOut(t *testing.T) {
	item := itemForTest()
	item.LimitedStack = 1
	fixture := newServiceFixture(t, item)

	result, err := fixture.service.Purchase(context.Background(), PurchaseParams{PlayerID: 7, CatalogItemID: 10, Rank: 1})
	if err != nil || result.LimitedUnitNumber == nil || *result.LimitedUnitNumber != 1 {
		t.Fatalf("unexpected result=%#v err=%v", result, err)
	}
	_, err = fixture.service.Purchase(context.Background(), PurchaseParams{PlayerID: 7, CatalogItemID: 10, Rank: 1})
	if !errors.Is(err, ErrOfferDisabled) {
		t.Fatalf("expected disabled sold out offer, got %v", err)
	}
}

// TestPurchaseRejectsInvisibleOffer verifies page rank policy before mutations.
func TestPurchaseRejectsInvisibleOffer(t *testing.T) {
	fixture := newServiceFixture(t, itemForTest())
	fixture.store.pages[0].MinRank = 5
	if err := fixture.service.Refresh(context.Background()); err != nil {
		t.Fatalf("refresh rank gate: %v", err)
	}

	_, err := fixture.service.Purchase(context.Background(), PurchaseParams{PlayerID: 7, CatalogItemID: 10, Rank: 1})
	if !errors.Is(err, ErrOfferNotVisible) || fixture.store.txCalls != 0 {
		t.Fatalf("expected visibility rejection, calls=%d err=%v", fixture.store.txCalls, err)
	}
}

// TestConcurrentPurchaseLastLimitedUnitAllowsOneWinner verifies serialized LTD reservation.
func TestConcurrentPurchaseLastLimitedUnitAllowsOneWinner(t *testing.T) {
	item := itemForTest()
	item.LimitedStack = 1
	fixture := newServiceFixture(t, item)
	start := make(chan struct{})
	errorsFound := make(chan error, 2)
	var wait sync.WaitGroup
	wait.Add(2)
	for range 2 {
		go func() {
			defer wait.Done()
			<-start
			_, err := fixture.service.Purchase(context.Background(), PurchaseParams{PlayerID: 7, CatalogItemID: 10, Rank: 1})
			errorsFound <- err
		}()
	}
	close(start)
	wait.Wait()
	close(errorsFound)

	successes := 0
	failures := 0
	for err := range errorsFound {
		if err == nil {
			successes++
		} else if errors.Is(err, ErrLimitedSoldOut) || errors.Is(err, ErrOfferDisabled) {
			failures++
		}
	}
	if successes != 1 || failures != 1 || len(fixture.furniture.calls) != 1 {
		t.Fatalf("expected one winner, successes=%d failures=%d grants=%d", successes, failures, len(fixture.furniture.calls))
	}
}

// TestBrowsingAndSanitizeUseRefreshedCache verifies K3 read behavior.
func TestBrowsingAndSanitizeUseRefreshedCache(t *testing.T) {
	fixture := newServiceFixture(t, itemForTest())
	fixture.store.sanitize = append(fixture.store.sanitize, furnituremodel.Definition{Name: "orphan_chair"})

	pages, err := fixture.service.Pages(context.Background(), 1, false)
	if err != nil || len(pages) != 1 {
		t.Fatalf("unexpected pages=%#v err=%v", pages, err)
	}
	_, items, err := fixture.service.Page(context.Background(), 1, 1, false)
	if err != nil || len(items) != 1 {
		t.Fatalf("unexpected items=%#v err=%v", items, err)
	}
	definitions, err := fixture.service.SanitizeList(context.Background())
	if err != nil || len(definitions) != 1 {
		t.Fatalf("unexpected definitions=%#v err=%v", definitions, err)
	}
}

// TestPagesRejectParentCycles verifies malformed page trees cannot loop.
func TestPagesRejectParentCycles(t *testing.T) {
	fixture := newServiceFixture(t, itemForTest())
	parentID := int64(1)
	fixture.store.pages[0].ParentID = &parentID
	if err := fixture.service.Refresh(context.Background()); err != nil {
		t.Fatalf("refresh cyclic page: %v", err)
	}

	pages, err := fixture.service.Pages(context.Background(), 1, false)
	if err != nil || len(pages) != 0 {
		t.Fatalf("expected cyclic page rejection, pages=%#v err=%v", pages, err)
	}
}
