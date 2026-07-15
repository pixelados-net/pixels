package service

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"

	"github.com/niflaot/pixels/internal/permission"
	catalogtrophy "github.com/niflaot/pixels/internal/realm/catalog/trophy"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
)

// TestPurchaseTrophyPersistsBuyerInscription verifies client data is filtered into server-owned trophy data.
func TestPurchaseTrophyPersistsBuyerInscription(t *testing.T) {
	item := itemForTest()
	fixture := newServiceFixture(t, item)
	fixture.furniture.definitions[0].InteractionType = "trophy"
	fixture.service.WithPlayers(purchasePlayerFinder{}).WithTrophies(catalogtrophy.New(nil))
	if err := fixture.service.Refresh(context.Background()); err != nil {
		t.Fatalf("refresh trophy definition: %v", err)
	}
	_, err := fixture.service.Purchase(context.Background(), PurchaseParams{PlayerID: 7, CatalogItemID: item.ID, ExtraData: "Winner\tmessage"})
	if err != nil {
		t.Fatalf("purchase trophy: %v", err)
	}
	if len(fixture.furniture.calls) != 1 || !strings.HasPrefix(fixture.furniture.calls[0].ExtraData, "demo\t") || !strings.HasSuffix(fixture.furniture.calls[0].ExtraData, "\tWinner message") {
		t.Fatalf("unexpected trophy grant %#v", fixture.furniture.calls)
	}
}

// TestPurchaseCreditsChargesAndGrants verifies a regular credits purchase.
func TestPurchaseCreditsChargesAndGrants(t *testing.T) {
	fixture := newServiceFixture(t, itemForTest())
	result, err := fixture.service.Purchase(context.Background(), PurchaseParams{PlayerID: 7, CatalogItemID: 10})
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

// TestPurchaseOverrideQuantityUsesFixedUnitPrice verifies targeted-offer quantities bypass catalog bulk policy.
func TestPurchaseOverrideQuantityUsesFixedUnitPrice(t *testing.T) {
	fixture := newServiceFixture(t, itemForTest())
	price := int64(7)
	result, err := fixture.service.Purchase(context.Background(), PurchaseParams{
		PlayerID: 7, CatalogItemID: 10, Amount: 2, OverrideCredits: &price,
	})
	if err != nil || len(result.GrantedItems) != 2 || len(fixture.currency.calls) != 1 ||
		fixture.currency.calls[0].Amount != -14 {
		t.Fatalf("result=%#v calls=%#v error=%v", result, fixture.currency.calls, err)
	}
}

// TestPurchasePairsGrantedTeleports verifies teleport offers create one durable pair.
func TestPurchasePairsGrantedTeleports(t *testing.T) {
	item := itemForTest()
	item.Amount = 2
	fixture := newServiceFixture(t, item)
	fixture.furniture.definitions[0].InteractionType = "teleport"
	if err := fixture.service.Refresh(context.Background()); err != nil {
		t.Fatalf("refresh teleport definition: %v", err)
	}

	result, err := fixture.service.Purchase(context.Background(), PurchaseParams{PlayerID: 7, CatalogItemID: 10})
	if err != nil {
		t.Fatalf("purchase teleport pair: %v", err)
	}
	if len(result.GrantedItems) != 2 || len(fixture.teleportPairs.pairs) != 1 || fixture.teleportPairs.pairs[0] != [2]int64{20, 21} {
		t.Fatalf("items=%#v pairs=%#v", result.GrantedItems, fixture.teleportPairs.pairs)
	}
}

// TestPurchaseRejectsUnpairedTeleportQuantity verifies malformed teleport offers fail atomically.
func TestPurchaseRejectsUnpairedTeleportQuantity(t *testing.T) {
	fixture := newServiceFixture(t, itemForTest())
	fixture.furniture.definitions[0].InteractionType = "teleport"
	if err := fixture.service.Refresh(context.Background()); err != nil {
		t.Fatalf("refresh teleport definition: %v", err)
	}

	_, err := fixture.service.Purchase(context.Background(), PurchaseParams{PlayerID: 7, CatalogItemID: 10})
	if !errors.Is(err, ErrTeleportPairing) || len(fixture.teleportPairs.pairs) != 0 {
		t.Fatalf("expected pairing rejection, pairs=%#v err=%v", fixture.teleportPairs.pairs, err)
	}
}

// TestPurchasePointsChargesConfiguredCurrency verifies an activity-points purchase.
func TestPurchasePointsChargesConfiguredCurrency(t *testing.T) {
	item := itemForTest()
	item.PointsType = 5
	item.CostCredits = 0
	item.CostPoints = 4
	fixture := newServiceFixture(t, item)

	result, err := fixture.service.Purchase(context.Background(), PurchaseParams{PlayerID: 7, CatalogItemID: 10})
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

	_, err := fixture.service.Purchase(context.Background(), PurchaseParams{PlayerID: 7, CatalogItemID: 10})
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

	result, err := fixture.service.Purchase(context.Background(), PurchaseParams{PlayerID: 7, CatalogItemID: 10})
	if err != nil || result.LimitedUnitNumber == nil || *result.LimitedUnitNumber != 1 {
		t.Fatalf("unexpected result=%#v err=%v", result, err)
	}
	_, err = fixture.service.Purchase(context.Background(), PurchaseParams{PlayerID: 7, CatalogItemID: 10})
	if !errors.Is(err, ErrOfferDisabled) {
		t.Fatalf("expected disabled sold out offer, got %v", err)
	}
}

// TestPurchaseRejectsInvisibleOffer verifies page permission policy before mutations.
func TestPurchaseRejectsInvisibleOffer(t *testing.T) {
	fixture := newServiceFixture(t, itemForTest())
	node := permission.RegisterNode("catalog.test.restricted", "")
	fixture.store.pages[0].RequiredNode = &node
	fixture.service.permissions = fixedChecker(false)
	if err := fixture.service.Refresh(context.Background()); err != nil {
		t.Fatalf("refresh permission gate: %v", err)
	}

	_, err := fixture.service.Purchase(context.Background(), PurchaseParams{PlayerID: 7, CatalogItemID: 10})
	if !errors.Is(err, ErrOfferNotVisible) || fixture.store.txCalls != 0 {
		t.Fatalf("expected visibility rejection, calls=%d err=%v", fixture.store.txCalls, err)
	}
}

// fixedChecker returns one configured permission result.
type fixedChecker bool

// HasPermission returns the configured permission result.
func (checker fixedChecker) HasPermission(context.Context, int64, permission.Node) (bool, error) {
	return bool(checker), nil
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
			_, err := fixture.service.Purchase(context.Background(), PurchaseParams{PlayerID: 7, CatalogItemID: 10})
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
