package core

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
)

// bonusFurniture supplies one Bonus Rare reward definition.
type bonusFurniture struct {
	furnitureservice.DefinitionGranter
	// definition stores the configured reward.
	definition furnituremodel.Definition
}

// FindDefinitionByID returns the configured reward definition.
func (furniture bonusFurniture) FindDefinitionByID(context.Context, int64) (furnituremodel.Definition, bool, error) {
	return furniture.definition, true, nil
}

// TestBonusRareInfoCalculatesProgress verifies below, equal, and above-threshold balances.
func TestBonusRareInfoCalculatesProgress(t *testing.T) {
	for _, test := range []struct {
		name      string
		balance   int64
		remaining int32
	}{{name: "below", balance: 45, remaining: 75}, {name: "equal", balance: 120}, {name: "above", balance: 180}} {
		t.Run(test.name, func(t *testing.T) {
			fixture := newCoreFixture(testTime())
			fixture.currencies.balance = test.balance
			fixture.service.options.BonusRareCurrencyType = 5
			fixture.service.options.BonusRareThreshold = 120
			fixture.service.options.BonusRareProductID = 91
			fixture.service.furniture = bonusFurniture{definition: furnituremodel.Definition{SpriteID: 712, PublicName: "Bonus Rare"}}
			info, err := fixture.service.BonusRareInfo(context.Background(), 7)
			if err != nil {
				t.Fatal(err)
			}
			if info.ProductType != "Bonus Rare" || info.ProductClassID != 712 || info.Threshold != 120 || info.Remaining != test.remaining {
				t.Fatalf("unexpected info %#v", info)
			}
		})
	}
}

// TestBonusRareInfoHandlesNeutralAndErrors verifies disabled reward and read failures.
func TestBonusRareInfoHandlesNeutralAndErrors(t *testing.T) {
	fixture := newCoreFixture(testTime())
	fixture.service.options.BonusRareThreshold = math.MaxInt64
	info, err := fixture.service.BonusRareInfo(context.Background(), 7)
	if err != nil || info.Threshold != math.MaxInt32 || info.ProductType != "" {
		t.Fatalf("info=%#v err=%v", info, err)
	}
	expected := errors.New("balance unavailable")
	fixture.currencies.balanceErr = expected
	if _, err = fixture.service.BonusRareInfo(context.Background(), 7); !errors.Is(err, expected) {
		t.Fatalf("expected balance failure, got %v", err)
	}
}

// BenchmarkBonusRareInfo measures the read-only hotel-view progress path.
func BenchmarkBonusRareInfo(b *testing.B) {
	fixture := newCoreFixture(testTime())
	fixture.currencies.balance = 60
	fixture.service.options.BonusRareThreshold = 120
	for b.Loop() {
		_, _ = fixture.service.BonusRareInfo(context.Background(), 7)
	}
}

// testTime returns the deterministic subscription fixture time.
func testTime() (result time.Time) {
	return time.Unix(1_700_000_000, 0)
}
