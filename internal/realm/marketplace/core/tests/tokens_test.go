package tests

import (
	"context"
	"errors"
	"testing"
	"time"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	marketcore "github.com/niflaot/pixels/internal/realm/marketplace/core"
)

// tokenStore provides configurable token persistence.
type tokenStore struct {
	// benchmarkStore supplies unrelated Marketplace persistence behavior.
	benchmarkStore
	// balance stores the current token count.
	balance int32
	// spendAllowed controls whether one token can be spent.
	spendAllowed bool
}

// AddTokens adds a package to the fixture balance.
func (store *tokenStore) AddTokens(_ context.Context, _ int64, amount int32) (int32, error) {
	store.balance += amount

	return store.balance, nil
}

// SpendToken spends one fixture token when allowed.
func (store *tokenStore) SpendToken(context.Context, int64) (bool, error) {
	if !store.spendAllowed || store.balance < 1 {
		return false, nil
	}
	store.balance--

	return true, nil
}

// tokenCurrency records the token-package credit deduction.
type tokenCurrency struct {
	// mutation stores the received currency mutation.
	mutation currencyservice.GrantParams
}

// Grant records a currency mutation.
func (currency *tokenCurrency) Grant(_ context.Context, params currencyservice.GrantParams) (int64, error) {
	currency.mutation = params

	return 99, nil
}

// tokenFurniture records Marketplace reservation attempts.
type tokenFurniture struct {
	// benchmarkFurniture supplies unrelated furniture behavior.
	benchmarkFurniture
	// reserved reports whether reservation was attempted.
	reserved bool
}

// ReserveForMarketplace records a furniture reservation attempt.
func (furniture *tokenFurniture) ReserveForMarketplace(context.Context, int64, int64) (furnituremodel.Item, furnituremodel.Definition, error) {
	furniture.reserved = true

	return furnituremodel.Item{}, furnituremodel.Definition{}, nil
}

// TestBuyTokensChargesConfiguredCredits verifies the explicit token purchase step.
func TestBuyTokensChargesConfiguredCredits(t *testing.T) {
	store := &tokenStore{}
	currency := &tokenCurrency{}
	marketplace := marketcore.New(marketcore.Options{Enabled: true, TokenCost: 3, TokenPackageSize: 5}, store, &benchmarkFurniture{}, currency, nil, nil)

	balance, err := marketplace.BuyTokens(context.Background(), 7)
	if err != nil || balance != 5 {
		t.Fatalf("balance=%d err=%v", balance, err)
	}
	if currency.mutation.PlayerID != 7 || currency.mutation.CurrencyType != -1 || currency.mutation.Amount != -3 || currency.mutation.ActorKind != currencyservice.ActorPlayer {
		t.Fatalf("unexpected currency mutation %#v", currency.mutation)
	}
}

// TestListWithoutTokenPreservesFurniture verifies a rejected post cannot reserve the item.
func TestListWithoutTokenPreservesFurniture(t *testing.T) {
	store := &tokenStore{}
	furniture := &tokenFurniture{}
	marketplace := marketcore.New(marketcore.Options{Enabled: true, MinimumPrice: 1, MaximumPrice: 100, OfferDuration: time.Hour}, store, furniture, nil, nil, nil)

	_, err := marketplace.List(context.Background(), 1, 28, 20)
	if !errors.Is(err, marketcore.ErrNoToken) {
		t.Fatalf("expected token error, got %v", err)
	}
	if furniture.reserved {
		t.Fatal("expected furniture to remain unreserved")
	}
}
