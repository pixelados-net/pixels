package service

import (
	"context"
	"errors"
	"testing"

	playereffect "github.com/niflaot/pixels/internal/realm/player/effect"
)

// effectManagerForTest records catalog-driven effect behavior.
type effectManagerForTest struct {
	// Manager supplies unused effect operations.
	playereffect.Manager
	// grants stores accepted effect ids.
	grants []int32
	// enables counts unexpected automatic selections.
	enables int
	// err fails grants when configured.
	err error
}

// Grant records one catalog effect charge.
func (manager *effectManagerForTest) Grant(_ context.Context, playerID int64, effectID int32, duration int32, source playereffect.Source) (playereffect.Effect, error) {
	if manager.err != nil {
		return playereffect.Effect{}, manager.err
	}
	manager.grants = append(manager.grants, effectID)
	return playereffect.Effect{PlayerID: playerID, ID: effectID, DurationSeconds: duration, RemainingCharges: 1}, nil
}

// Enable records an unexpected catalog selection.
func (manager *effectManagerForTest) Enable(context.Context, int64, int32) error {
	manager.enables++
	return nil
}

// TestPurchasePureEffectGrantsNoFurniture verifies effect-only catalog offers.
func TestPurchasePureEffectGrantsNoFurniture(t *testing.T) {
	effectID := int32(101)
	item := itemForTest()
	item.DefinitionID = 0
	item.Amount = 0
	item.GrantsEffectID = &effectID
	item.GrantsEffectDurationSeconds = 86400
	fixture := newServiceFixture(t, item)
	effects := &effectManagerForTest{}
	fixture.service.WithEffects(effects)

	result, err := fixture.service.Purchase(context.Background(), PurchaseParams{PlayerID: 7, CatalogItemID: item.ID})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.GrantedItems) != 0 || len(fixture.furniture.calls) != 0 || len(effects.grants) != 1 || effects.grants[0] != effectID || effects.enables != 0 {
		t.Fatalf("result=%#v furniture=%#v effects=%#v", result, fixture.furniture.calls, effects)
	}
}

// TestPurchaseMixedRewardGrantsFurnitureAndEffect verifies the shared pipeline.
func TestPurchaseMixedRewardGrantsFurnitureAndEffect(t *testing.T) {
	effectID := int32(103)
	item := itemForTest()
	item.GrantsEffectID = &effectID
	fixture := newServiceFixture(t, item)
	effects := &effectManagerForTest{}
	fixture.service.WithEffects(effects)

	result, err := fixture.service.Purchase(context.Background(), PurchaseParams{PlayerID: 7, CatalogItemID: item.ID})
	if err != nil || len(result.GrantedItems) != 1 || len(effects.grants) != 1 || result.GrantedEffectID == nil || *result.GrantedEffectID != effectID {
		t.Fatalf("result=%#v effects=%#v err=%v", result, effects, err)
	}
}

// TestPurchaseEffectFailureAbortsCompletion verifies grant failures fail commerce.
func TestPurchaseEffectFailureAbortsCompletion(t *testing.T) {
	effectID := int32(103)
	item := itemForTest()
	item.GrantsEffectID = &effectID
	fixture := newServiceFixture(t, item)
	want := errors.New("effect unavailable")
	fixture.service.WithEffects(&effectManagerForTest{err: want})

	if _, err := fixture.service.Purchase(context.Background(), PurchaseParams{PlayerID: 7, CatalogItemID: item.ID}); !errors.Is(err, want) {
		t.Fatalf("expected effect failure, got %v", err)
	}
}

// TestPurchaseClubEffectUsesExistingVisibilityGate verifies no parallel effect policy.
func TestPurchaseClubEffectUsesExistingVisibilityGate(t *testing.T) {
	effectID := int32(201)
	item := itemForTest()
	item.DefinitionID = 0
	item.Amount = 0
	item.GrantsEffectID = &effectID
	item.ClubOnly = true
	fixture := newServiceFixture(t, item)
	effects := &effectManagerForTest{}
	fixture.service.WithEffects(effects)

	_, err := fixture.service.Purchase(context.Background(), PurchaseParams{PlayerID: 7, CatalogItemID: item.ID})
	if !errors.Is(err, ErrOfferNotVisible) || len(effects.grants) != 0 {
		t.Fatalf("expected visibility rejection, grants=%v err=%v", effects.grants, err)
	}
}
