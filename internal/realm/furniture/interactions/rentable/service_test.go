package rentable

import (
	"context"
	"errors"
	"testing"
	"time"

	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
)

// rentableStoreFixture records guarded rental mutations.
type rentableStoreFixture struct {
	state    State
	found    bool
	rentOK   bool
	cancelOK bool
	buyoutOK bool
}

// FindRoomSpace returns the fixture state.
func (store *rentableStoreFixture) FindRoomSpace(context.Context, int64) (State, bool, error) {
	return store.state, store.found, nil
}

// FindItem returns the fixture state.
func (store *rentableStoreFixture) FindItem(context.Context, int64) (State, bool, error) {
	return store.state, store.found, nil
}

// Rent records one guarded mutation.
func (store *rentableStoreFixture) Rent(_ context.Context, _ int64, renterID int64, price int32, durationSeconds int64) (State, bool, error) {
	expires := time.Unix(100, 0).Add(time.Duration(durationSeconds) * time.Second)
	store.state.RenterPlayerID = &renterID
	store.state.ExpiresAt = &expires
	store.state.PriceCredits = price
	return store.state, store.rentOK, nil
}

// Cancel returns the configured guarded result.
func (store *rentableStoreFixture) Cancel(context.Context, int64, int64) (bool, error) {
	return store.cancelOK, nil
}

// Buyout returns the configured guarded result.
func (store *rentableStoreFixture) Buyout(context.Context, int64, int64) (bool, error) {
	return store.buyoutOK, nil
}

// WithinTransaction executes one fixture transaction.
func (store *rentableStoreFixture) WithinTransaction(ctx context.Context, work func(context.Context) error) error {
	return work(ctx)
}

// currencyFixture records signed credit mutations.
type currencyFixture struct {
	grants []currencyservice.GrantParams
	err    error
}

// Grant records one balance mutation.
func (currency *currencyFixture) Grant(_ context.Context, params currencyservice.GrantParams) (int64, error) {
	currency.grants = append(currency.grants, params)
	return 100 + params.Amount, currency.err
}

// TestRentChargesAndExtendsAtomically verifies the normal rental path.
func TestRentChargesAndExtendsAtomically(t *testing.T) {
	store := &rentableStoreFixture{state: State{ItemID: 9}, found: true, rentOK: true}
	currency := &currencyFixture{}
	service := New(Config{Duration: time.Hour, PriceCredits: 10}, store, currency)
	service.now = func() time.Time { return time.Unix(100, 0) }
	state, err := service.Rent(context.Background(), 9, 7)
	if err != nil || state.RenterPlayerID == nil || *state.RenterPlayerID != 7 || len(currency.grants) != 1 || currency.grants[0].Amount != -10 {
		t.Fatalf("state=%+v grants=%+v err=%v", state, currency.grants, err)
	}
}

// TestRentRejectsAnotherActiveRenterBeforeCharging verifies conflict ordering.
func TestRentRejectsAnotherActiveRenterBeforeCharging(t *testing.T) {
	renter := int64(8)
	expires := time.Unix(200, 0)
	store := &rentableStoreFixture{state: State{ItemID: 9, RenterPlayerID: &renter, ExpiresAt: &expires}, found: true, rentOK: true}
	currency := &currencyFixture{}
	service := New(Config{PriceCredits: 10}, store, currency)
	service.now = func() time.Time { return time.Unix(100, 0) }
	_, err := service.Rent(context.Background(), 9, 7)
	if !errors.Is(err, ErrUnavailable) || len(currency.grants) != 0 {
		t.Fatalf("grants=%+v err=%v", currency.grants, err)
	}
}

// TestRentPropagatesChargeFailureWithoutMutation verifies transactional debit failure.
func TestRentPropagatesChargeFailureWithoutMutation(t *testing.T) {
	expected := errors.New("insufficient credits")
	store := &rentableStoreFixture{state: State{ItemID: 9}, found: true, rentOK: true}
	service := New(Config{PriceCredits: 10}, store, &currencyFixture{err: expected})
	if _, err := service.Rent(context.Background(), 9, 7); !errors.Is(err, expected) || store.state.RenterPlayerID != nil {
		t.Fatalf("state=%+v err=%v", store.state, err)
	}
}

// TestCancelAndBuyoutRequireActiveRenter verifies guarded terminal transitions.
func TestCancelAndBuyoutRequireActiveRenter(t *testing.T) {
	renter := int64(7)
	expires := time.Now().Add(time.Hour)
	store := &rentableStoreFixture{state: State{ItemID: 9, RenterPlayerID: &renter, ExpiresAt: &expires}, found: true}
	service := New(Config{BuyoutCredits: 50}, store, &currencyFixture{})
	if err := service.Cancel(context.Background(), 9, 7); !errors.Is(err, ErrNotRenter) {
		t.Fatalf("cancel err=%v", err)
	}
	if err := service.Buyout(context.Background(), 9, 7); !errors.Is(err, ErrNotRenter) {
		t.Fatalf("buyout err=%v", err)
	}
	store.cancelOK, store.buyoutOK = true, true
	if err := service.Cancel(context.Background(), 9, 7); err != nil {
		t.Fatal(err)
	}
	if err := service.Buyout(context.Background(), 9, 7); err != nil {
		t.Fatal(err)
	}
}
