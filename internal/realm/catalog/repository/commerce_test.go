package repository

import (
	"context"
	"strings"
	"testing"
	"time"

	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
)

// TestLogPurchaseLinksGrantedFurniture verifies purchase provenance persistence.
func TestLogPurchaseLinksGrantedFurniture(t *testing.T) {
	executor := &fakeExecutor{row: fakeRow{values: []any{int64(42)}}}
	item := catalogmodel.Item{PointsType: -1}
	item.ID = 3
	err := newRepository(executor).LogPurchase(context.Background(), 7, item, 2, 24, 0, []int64{91, 92})
	if err != nil || !strings.Contains(executor.query, "catalog_purchase_items") || len(executor.arguments) != 2 {
		t.Fatalf("query=%q arguments=%#v error=%v", executor.query, executor.arguments, err)
	}
}

// TestCreditsSpentBetweenUsesBothPeriodBoundaries verifies cycle isolation.
func TestCreditsSpentBetweenUsesBothPeriodBoundaries(t *testing.T) {
	executor := &fakeExecutor{row: fakeRow{values: []any{int64(35)}}}
	after := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	through := after.Add(31 * 24 * time.Hour)
	amount, err := newRepository(executor).CreditsSpentBetween(context.Background(), 7, after, through)
	if err != nil || amount != 35 || len(executor.arguments) != 3 || !strings.Contains(executor.query, "purchased_at<=$3") {
		t.Fatalf("amount=%d query=%q arguments=%#v error=%v", amount, executor.query, executor.arguments, err)
	}
}
