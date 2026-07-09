// Package holder contains the live player currency capability.
package holder

import (
	"context"

	currencymodel "github.com/niflaot/pixels/internal/realm/inventory/currency/model"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
)

// Holder identifies one player's currency capability without caching durable balances.
type Holder struct {
	// playerID identifies the wallet owner.
	playerID int64
}

// New creates a player currency holder.
func New(playerID int64) *Holder {
	return &Holder{playerID: playerID}
}

// PlayerID returns the wallet owner.
func (holder *Holder) PlayerID() int64 {
	return holder.playerID
}

// Wallet reads the current durable wallet through a narrow reader contract.
func (holder *Holder) Wallet(ctx context.Context, currencies currencyservice.Reader) ([]currencymodel.Balance, error) {
	return currencies.Wallet(ctx, holder.playerID)
}

// Balance reads one current durable balance through a narrow reader contract.
func (holder *Holder) Balance(ctx context.Context, currencies currencyservice.Reader, currencyType int32) (int64, error) {
	return currencies.Balance(ctx, holder.playerID, currencyType)
}
