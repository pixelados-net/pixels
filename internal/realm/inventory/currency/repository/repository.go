package repository

import (
	"context"
	"fmt"
	"math"

	"github.com/jackc/pgx/v5"
	currencymodel "github.com/niflaot/pixels/internal/realm/inventory/currency/model"
	"github.com/niflaot/pixels/pkg/postgres"
)

// txRunner runs currency work inside one transaction.
type txRunner func(context.Context, func(context.Context, postgres.Executor) error) error

// Repository reads and writes currency persistence records.
type Repository struct {
	// executor runs read-only PostgreSQL queries.
	executor postgres.Executor

	// withinTx runs atomic currency mutations.
	withinTx txRunner
}

// New creates a currency repository.
func New(pool *postgres.Pool) *Repository {
	return &Repository{
		executor: pool,
		withinTx: func(ctx context.Context, work func(context.Context, postgres.Executor) error) error {
			return postgres.WithinTx(ctx, pool, func(ctx context.Context, tx pgx.Tx) error {
				return work(ctx, tx)
			})
		},
	}
}

// FindBalance finds one player currency balance.
func (repository *Repository) FindBalance(ctx context.Context, playerID int64, currencyType int32) (currencymodel.Balance, bool, error) {
	return findBalance(ctx, repository.executor, playerID, currencyType)
}

// ListBalances lists one player's stored currency balances.
func (repository *Repository) ListBalances(ctx context.Context, playerID int64) ([]currencymodel.Balance, error) {
	rows, err := repository.executor.Query(ctx, `
		select player_id, currency_type, amount, updated_at, version
		from player_currencies
		where player_id = $1
		order by currency_type`, playerID)
	if err != nil {
		return nil, fmt.Errorf("list player %d currency balances: %w", playerID, err)
	}
	defer rows.Close()

	balances := make([]currencymodel.Balance, 0)
	for rows.Next() {
		var balance currencymodel.Balance
		if err := rows.Scan(&balance.PlayerID, &balance.CurrencyType, &balance.Amount, &balance.UpdatedAt, &balance.Version); err != nil {
			return nil, fmt.Errorf("scan player %d currency balance: %w", playerID, err)
		}
		balances = append(balances, balance)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate player %d currency balances: %w", playerID, err)
	}

	return balances, nil
}

// Grant applies a signed delta and optional ledger entry atomically.
func (repository *Repository) Grant(ctx context.Context, mutation Mutation) (Result, error) {
	return repository.mutate(ctx, mutation, false)
}

// Set replaces a balance and writes an optional ledger entry atomically.
func (repository *Repository) Set(ctx context.Context, mutation Mutation) (Result, error) {
	return repository.mutate(ctx, mutation, true)
}

// mutate serializes and persists one currency mutation.
func (repository *Repository) mutate(ctx context.Context, mutation Mutation, absolute bool) (Result, error) {
	var result Result
	err := repository.withinTx(ctx, func(ctx context.Context, executor postgres.Executor) error {
		if err := lockBalance(ctx, executor, mutation.PlayerID, mutation.CurrencyType); err != nil {
			return err
		}

		current, found, err := findBalance(ctx, executor, mutation.PlayerID, mutation.CurrencyType)
		if err != nil {
			return err
		}
		previous := int64(0)
		if found {
			previous = current.Amount
		}

		if !absolute && mutation.Amount > 0 && previous > math.MaxInt64-mutation.Amount {
			return ErrBalanceOverflow
		}
		next := previous + mutation.Amount
		if absolute {
			next = mutation.Amount
		}
		if next < 0 {
			return ErrInsufficientBalance
		}

		result.Balance, err = upsertBalance(ctx, executor, mutation.PlayerID, mutation.CurrencyType, next)
		if err != nil {
			return err
		}
		result.Delta = next - previous
		if mutation.Ledger {
			return insertLedger(ctx, executor, ledgerEntry(mutation, result.Delta, next))
		}

		return nil
	})
	if err != nil {
		return Result{}, err
	}

	return result, nil
}
