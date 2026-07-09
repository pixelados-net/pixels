package service

import (
	"context"
	"errors"
	"testing"

	"github.com/niflaot/pixels/internal/realm/inventory/currency"
	currencymodel "github.com/niflaot/pixels/internal/realm/inventory/currency/model"
	currencyrepo "github.com/niflaot/pixels/internal/realm/inventory/currency/repository"
)

// TestWalletProjectsConfiguredTypes verifies missing balances are represented as zero.
func TestWalletProjectsConfiguredTypes(t *testing.T) {
	store := &fakeStore{balances: []currencymodel.Balance{{PlayerID: 7, CurrencyType: 5, Amount: 9}}}
	service := newTestService(t, store)

	wallet, err := service.Wallet(context.Background(), 7)
	if err != nil {
		t.Fatalf("wallet: %v", err)
	}
	if len(wallet) != 3 || wallet[0].CurrencyType != -1 || wallet[0].Amount != 0 || wallet[2].Amount != 9 {
		t.Fatalf("unexpected wallet %#v", wallet)
	}
}

// TestBalanceAndTypesReadConfiguredState verifies focused reader behavior.
func TestBalanceAndTypesReadConfiguredState(t *testing.T) {
	store := &fakeStore{balances: []currencymodel.Balance{{PlayerID: 7, CurrencyType: -1, Amount: 25}}}
	service := newTestService(t, store)

	amount, err := service.Balance(context.Background(), 7, -1)
	if err != nil || amount != 25 {
		t.Fatalf("unexpected balance=%d err=%v", amount, err)
	}
	types, err := service.Types(context.Background())
	if err != nil || len(types) != 3 {
		t.Fatalf("unexpected types=%#v err=%v", types, err)
	}
}

// TestGrantValidatesAndPersists verifies grant behavior and ledger selection.
func TestGrantValidatesAndPersists(t *testing.T) {
	store := &fakeStore{grantBalance: currencymodel.Balance{Amount: 15}}
	service := newTestService(t, store)

	amount, err := service.Grant(context.Background(), GrantParams{
		PlayerID: 7, CurrencyType: -1, Amount: 5, ActorKind: ActorSystem,
	})
	if err != nil {
		t.Fatalf("grant: %v", err)
	}
	if amount != 15 || !store.mutation.Ledger || store.mutation.Amount != 5 {
		t.Fatalf("unexpected grant amount=%d mutation=%#v", amount, store.mutation)
	}

	_, err = service.Grant(context.Background(), GrantParams{
		PlayerID: 7, CurrencyType: 0, Amount: 1, ActorKind: ActorSystem,
	})
	if err != nil {
		t.Fatalf("grant duckets: %v", err)
	}
	if store.mutation.Ledger {
		t.Fatal("expected duckets mutation without ledger")
	}
}

// TestGrantMapsInsufficientBalance verifies repository conflicts become domain errors.
func TestGrantMapsInsufficientBalance(t *testing.T) {
	store := &fakeStore{grantErr: currencyrepo.ErrInsufficientBalance}
	service := newTestService(t, store)

	_, err := service.Grant(context.Background(), GrantParams{
		PlayerID: 7, CurrencyType: -1, Amount: -1, ActorKind: ActorPlayer,
	})
	if !errors.Is(err, ErrInsufficientBalance) {
		t.Fatalf("expected insufficient balance, got %v", err)
	}
}

// TestSetValidatesAbsoluteAmount verifies set validation and persistence.
func TestSetValidatesAbsoluteAmount(t *testing.T) {
	store := &fakeStore{setBalance: currencymodel.Balance{Amount: 3}}
	service := newTestService(t, store)

	amount, err := service.Set(context.Background(), SetParams{
		PlayerID: 7, CurrencyType: -1, Amount: 3, ActorKind: ActorAdmin,
	})
	if err != nil {
		t.Fatalf("set: %v", err)
	}
	if amount != 3 || store.mutation.Amount != 3 || !store.mutation.Ledger {
		t.Fatalf("unexpected set amount=%d mutation=%#v", amount, store.mutation)
	}

	_, err = service.Set(context.Background(), SetParams{
		PlayerID: 7, CurrencyType: -1, Amount: -1, ActorKind: ActorAdmin,
	})
	if !errors.Is(err, ErrInvalidAmount) {
		t.Fatalf("expected invalid amount, got %v", err)
	}
}

// TestServiceRejectsInvalidInputs verifies identity, amount, type, and actor validation.
func TestServiceRejectsInvalidInputs(t *testing.T) {
	service := newTestService(t, &fakeStore{})
	cases := []struct {
		name   string
		params GrantParams
		target error
	}{
		{name: "player", params: GrantParams{CurrencyType: -1, Amount: 1, ActorKind: ActorSystem}, target: ErrInvalidPlayerID},
		{name: "type", params: GrantParams{PlayerID: 7, CurrencyType: 99, Amount: 1, ActorKind: ActorSystem}, target: ErrInvalidCurrencyType},
		{name: "amount", params: GrantParams{PlayerID: 7, CurrencyType: -1, ActorKind: ActorSystem}, target: ErrInvalidAmount},
		{name: "actor", params: GrantParams{PlayerID: 7, CurrencyType: -1, Amount: 1, ActorKind: "other"}, target: ErrInvalidActor},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			_, err := service.Grant(context.Background(), test.params)
			if !errors.Is(err, test.target) {
				t.Fatalf("expected %v, got %v", test.target, err)
			}
		})
	}
}

// newTestService creates a currency service test subject.
func newTestService(t *testing.T, store *fakeStore) *Service {
	t.Helper()
	catalog, err := currency.NewCatalog([]currencymodel.Definition{
		{Type: -1, Key: "credits"}, {Type: 0, Key: "duckets"}, {Type: 5, Key: "diamonds"},
	}, []int32{-1})
	if err != nil {
		t.Fatalf("new catalog: %v", err)
	}

	return New(store, catalog)
}

// fakeStore records currency service persistence calls.
type fakeStore struct {
	// balances stores wallet results.
	balances []currencymodel.Balance
	// grantBalance stores grant results.
	grantBalance currencymodel.Balance
	// setBalance stores set results.
	setBalance currencymodel.Balance
	// grantErr stores grant errors.
	grantErr error
	// mutation stores the last mutation.
	mutation currencyrepo.Mutation
}

// FindBalance finds a fake balance.
func (store *fakeStore) FindBalance(context.Context, int64, int32) (currencymodel.Balance, bool, error) {
	if len(store.balances) == 0 {
		return currencymodel.Balance{}, false, nil
	}
	return store.balances[0], true, nil
}

// ListBalances lists fake balances.
func (store *fakeStore) ListBalances(context.Context, int64) ([]currencymodel.Balance, error) {
	return store.balances, nil
}

// Grant records a fake grant.
func (store *fakeStore) Grant(_ context.Context, mutation currencyrepo.Mutation) (currencymodel.Balance, error) {
	store.mutation = mutation
	return store.grantBalance, store.grantErr
}

// Set records a fake absolute set.
func (store *fakeStore) Set(_ context.Context, mutation currencyrepo.Mutation) (currencymodel.Balance, error) {
	store.mutation = mutation
	return store.setBalance, nil
}
