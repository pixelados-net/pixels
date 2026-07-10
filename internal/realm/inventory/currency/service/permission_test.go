package service

import (
	"context"
	"errors"
	"testing"

	"github.com/niflaot/pixels/internal/permission"
	"github.com/niflaot/pixels/internal/realm/inventory/currency"
	currencymodel "github.com/niflaot/pixels/internal/realm/inventory/currency/model"
)

// fakeChecker returns one permission decision.
type fakeChecker struct {
	// allowed stores the returned decision.
	allowed bool
	// err stores the returned failure.
	err error
	// calls stores permission lookup count.
	calls int
}

// HasPermission resolves one fixture permission decision.
func (checker *fakeChecker) HasPermission(_ context.Context, _ int64, node permission.Node) (bool, error) {
	checker.calls++
	if node != currency.InfiniteBalance {
		return false, errors.New("unexpected permission node")
	}
	return checker.allowed, checker.err
}

// TestInfiniteBalanceSkipsPlayerDeductions verifies privileged purchases remain free.
func TestInfiniteBalanceSkipsPlayerDeductions(t *testing.T) {
	store := &fakeStore{balances: []currencymodel.Balance{{PlayerID: 7, CurrencyType: -1, Amount: 25}}}
	checker := &fakeChecker{allowed: true}
	service := newTestService(t, store)
	service.permissions = checker

	amount, err := service.Grant(context.Background(), GrantParams{PlayerID: 7, CurrencyType: -1, Amount: -5, ActorKind: ActorPlayer})
	if err != nil || amount != 25 {
		t.Fatalf("unexpected balance=%d err=%v", amount, err)
	}
	if store.mutation.Amount != 0 || checker.calls != 1 {
		t.Fatalf("expected no persisted deduction, mutation=%#v calls=%d", store.mutation, checker.calls)
	}
}

// TestInfiniteBalanceDoesNotBypassAdminDeductions verifies actor scoping.
func TestInfiniteBalanceDoesNotBypassAdminDeductions(t *testing.T) {
	store := &fakeStore{grantBalance: currencymodel.Balance{Amount: 20}}
	checker := &fakeChecker{allowed: true}
	service := newTestService(t, store)
	service.permissions = checker

	amount, err := service.Grant(context.Background(), GrantParams{PlayerID: 7, CurrencyType: -1, Amount: -5, ActorKind: ActorAdmin})
	if err != nil || amount != 20 || store.mutation.Amount != -5 || checker.calls != 0 {
		t.Fatalf("unexpected balance=%d mutation=%#v calls=%d err=%v", amount, store.mutation, checker.calls, err)
	}
}

// TestInfiniteBalancePropagatesPermissionFailures verifies purchase checks fail closed.
func TestInfiniteBalancePropagatesPermissionFailures(t *testing.T) {
	failure := errors.New("permission unavailable")
	service := newTestService(t, &fakeStore{})
	service.permissions = &fakeChecker{err: failure}

	_, err := service.Grant(context.Background(), GrantParams{PlayerID: 7, CurrencyType: -1, Amount: -5, ActorKind: ActorPlayer})
	if !errors.Is(err, failure) {
		t.Fatalf("expected permission failure, got %v", err)
	}
}

// BenchmarkInfiniteBalanceDeduction measures permission-aware free purchases.
func BenchmarkInfiniteBalanceDeduction(b *testing.B) {
	store := &fakeStore{balances: []currencymodel.Balance{{PlayerID: 7, CurrencyType: -1, Amount: 25}}}
	service := newTestService(b, store)
	service.permissions = &fakeChecker{allowed: true}
	params := GrantParams{PlayerID: 7, CurrencyType: -1, Amount: -5, ActorKind: ActorPlayer}
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		amount, err := service.Grant(ctx, params)
		if err != nil || amount != 25 {
			b.Fatalf("unexpected amount=%d err=%v", amount, err)
		}
	}
}
