package service

import (
	"context"
	"errors"
	"testing"

	"github.com/niflaot/pixels/internal/realm/inventory/currency"
	currencychanged "github.com/niflaot/pixels/internal/realm/inventory/currency/events/changed"
	currencymodel "github.com/niflaot/pixels/internal/realm/inventory/currency/model"
	currencyrepo "github.com/niflaot/pixels/internal/realm/inventory/currency/repository"
	"github.com/niflaot/pixels/pkg/bus"
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

// TestGrantPublishesCommittedChange verifies mutation events use repository results.
func TestGrantPublishesCommittedChange(t *testing.T) {
	store := &fakeStore{grantBalance: currencymodel.Balance{PlayerID: 7, CurrencyType: 5, Amount: 15}}
	publisher := &fakePublisher{}
	service := newTestServiceWithPublisher(t, store, publisher)

	_, err := service.Grant(context.Background(), GrantParams{
		PlayerID: 7, CurrencyType: 5, Amount: 5, ActorKind: ActorSystem,
	})
	if err != nil {
		t.Fatalf("grant: %v", err)
	}
	if len(publisher.events) != 1 || publisher.events[0].Name != currencychanged.Name {
		t.Fatalf("unexpected events %#v", publisher.events)
	}
	payload, ok := publisher.events[0].Payload.(currencychanged.Payload)
	if !ok || payload.Amount != 15 || payload.Delta != 5 || payload.PlayerID != 7 {
		t.Fatalf("unexpected payload %#v", publisher.events[0].Payload)
	}
}

// TestGrantKeepsCommittedResultWhenProjectionFails verifies event side effects cannot undo persistence.
func TestGrantKeepsCommittedResultWhenProjectionFails(t *testing.T) {
	store := &fakeStore{grantBalance: currencymodel.Balance{PlayerID: 7, CurrencyType: 5, Amount: 15}}
	publisher := &fakePublisher{err: errors.New("projection failed")}
	service := newTestServiceWithPublisher(t, store, publisher)

	amount, err := service.Grant(context.Background(), GrantParams{
		PlayerID: 7, CurrencyType: 5, Amount: 5, ActorKind: ActorSystem,
	})
	if err != nil || amount != 15 {
		t.Fatalf("unexpected committed amount=%d err=%v", amount, err)
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
	return newTestServiceWithPublisher(t, store, nil)
}

// newTestServiceWithPublisher creates a currency service with event publishing.
func newTestServiceWithPublisher(t *testing.T, store *fakeStore, publisher bus.Publisher) *Service {
	t.Helper()
	catalog, err := currency.NewCatalog([]currencymodel.Definition{
		{Type: -1, Key: "credits"}, {Type: 0, Key: "duckets"}, {Type: 5, Key: "diamonds"},
	}, []int32{-1})
	if err != nil {
		t.Fatalf("new catalog: %v", err)
	}

	return New(store, catalog, publisher, nil)
}

// fakePublisher records currency service events.
type fakePublisher struct {
	// events stores published events.
	events []bus.Event
	// err stores a publication failure.
	err error
}

// Publish records one event.
func (publisher *fakePublisher) Publish(_ context.Context, event bus.Event) error {
	publisher.events = append(publisher.events, event)
	return publisher.err
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
func (store *fakeStore) Grant(_ context.Context, mutation currencyrepo.Mutation) (currencyrepo.Result, error) {
	store.mutation = mutation
	return currencyrepo.Result{Balance: store.grantBalance, Delta: mutation.Amount}, store.grantErr
}

// Set records a fake absolute set.
func (store *fakeStore) Set(_ context.Context, mutation currencyrepo.Mutation) (currencyrepo.Result, error) {
	store.mutation = mutation
	return currencyrepo.Result{Balance: store.setBalance, Delta: -7}, nil
}
