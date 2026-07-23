package repository

import (
	"context"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

// TestLimitedUnitLifecycleUsesAtomicStatements verifies numbered LTD persistence.
func TestLimitedUnitLifecycleUsesAtomicStatements(t *testing.T) {
	executor := &fakeExecutor{}
	repository := newRepository(executor)
	if err := repository.CreateLimitedUnits(context.Background(), 2, 10); err != nil {
		t.Fatalf("create limited units: %v", err)
	}
	if !strings.Contains(executor.query, "generate_series") {
		t.Fatalf("unexpected create query %q", executor.query)
	}

	executor.row = fakeRow{values: []any{int32(1)}}
	number, reserved, err := repository.ReserveLimitedUnit(context.Background(), 2, 7)
	if err != nil || !reserved || number != 1 || !strings.Contains(executor.query, "skip locked") {
		t.Fatalf("unexpected number=%d reserved=%v query=%q err=%v", number, reserved, executor.query, err)
	}

	executor.row = fakeRow{values: []any{true}}
	completed, err := repository.CompleteLimitedUnit(context.Background(), 2, 1, 7, 20)
	if err != nil || !completed || !strings.Contains(executor.query, "limited_sells=limited_sells+1") {
		t.Fatalf("unexpected completed=%v query=%q err=%v", completed, executor.query, err)
	}
}

// TestReserveLimitedUnitReportsSoldOut verifies empty LTD inventory.
func TestReserveLimitedUnitReportsSoldOut(t *testing.T) {
	executor := &fakeExecutor{row: fakeRow{err: pgx.ErrNoRows}}
	_, reserved, err := newRepository(executor).ReserveLimitedUnit(context.Background(), 2, 7)
	if err != nil || reserved {
		t.Fatalf("expected sold out reservation, reserved=%v err=%v", reserved, err)
	}
}
