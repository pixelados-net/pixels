// Package repository contains PostgreSQL currency persistence.
package repository

import (
	"context"

	currencymodel "github.com/niflaot/pixels/internal/realm/inventory/currency/model"
)

// Store reads and atomically mutates currency balances.
type Store interface {
	// FindBalance finds one player currency balance.
	FindBalance(ctx context.Context, playerID int64, currencyType int32) (currencymodel.Balance, bool, error)

	// ListBalances lists one player's stored currency balances.
	ListBalances(ctx context.Context, playerID int64) ([]currencymodel.Balance, error)

	// Grant applies a signed delta and optional ledger entry atomically.
	Grant(ctx context.Context, mutation Mutation) (Result, error)

	// Set replaces a balance and writes an optional ledger entry atomically.
	Set(ctx context.Context, mutation Mutation) (Result, error)
}

// storeAssertion verifies Repository implements Store.
var storeAssertion Store = (*Repository)(nil)
