package holder

import (
	"context"
	"testing"

	currencymodel "github.com/niflaot/pixels/internal/realm/inventory/currency/model"
)

// TestHolderReadsOwnedWallet verifies composed wallet access.
func TestHolderReadsOwnedWallet(t *testing.T) {
	holder := New(7)
	reader := fakeReader{}
	wallet, err := holder.Wallet(context.Background(), reader)
	if err != nil {
		t.Fatalf("wallet: %v", err)
	}
	if holder.PlayerID() != 7 || len(wallet) != 1 || wallet[0].PlayerID != 7 {
		t.Fatalf("unexpected holder=%d wallet=%#v", holder.PlayerID(), wallet)
	}
	balance, err := holder.Balance(context.Background(), reader, -1)
	if err != nil || balance != 10 {
		t.Fatalf("unexpected balance=%d err=%v", balance, err)
	}
}

// fakeReader returns holder test balances.
type fakeReader struct{}

// Wallet returns one fake wallet balance.
func (fakeReader) Wallet(_ context.Context, playerID int64) ([]currencymodel.Balance, error) {
	return []currencymodel.Balance{{PlayerID: playerID, CurrencyType: -1, Amount: 10}}, nil
}

// Balance returns one fake balance.
func (fakeReader) Balance(context.Context, int64, int32) (int64, error) {
	return 10, nil
}

// Types returns no fake definitions.
func (fakeReader) Types(context.Context) ([]currencymodel.Definition, error) {
	return nil, nil
}
