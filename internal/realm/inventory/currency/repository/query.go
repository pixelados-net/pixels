package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	currencymodel "github.com/niflaot/pixels/internal/realm/inventory/currency/model"
	"github.com/niflaot/pixels/pkg/postgres"
)

// lockBalance serializes mutations for one player currency across server instances.
func lockBalance(ctx context.Context, executor postgres.Executor, playerID int64, currencyType int32) error {
	key := fmt.Sprintf("%d:%d", playerID, currencyType)
	if _, err := executor.Exec(ctx, "select pg_advisory_xact_lock(hashtextextended($1, 0))", key); err != nil {
		return fmt.Errorf("lock player %d currency %d: %w", playerID, currencyType, err)
	}

	return nil
}

// findBalance finds one balance using an executor.
func findBalance(ctx context.Context, executor postgres.Executor, playerID int64, currencyType int32) (currencymodel.Balance, bool, error) {
	var balance currencymodel.Balance
	err := executor.QueryRow(ctx, `
		select player_id, currency_type, amount, updated_at, version
		from player_currencies
		where player_id = $1 and currency_type = $2`, playerID, currencyType).Scan(
		&balance.PlayerID,
		&balance.CurrencyType,
		&balance.Amount,
		&balance.UpdatedAt,
		&balance.Version,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return currencymodel.Balance{}, false, nil
	}
	if err != nil {
		return currencymodel.Balance{}, false, fmt.Errorf("find player %d currency %d: %w", playerID, currencyType, err)
	}

	return balance, true, nil
}

// upsertBalance stores and returns one absolute balance.
func upsertBalance(ctx context.Context, executor postgres.Executor, playerID int64, currencyType int32, amount int64) (currencymodel.Balance, error) {
	var balance currencymodel.Balance
	err := executor.QueryRow(ctx, `
		insert into player_currencies (player_id, currency_type, amount)
		values ($1, $2, $3)
		on conflict (player_id, currency_type) do update
		set amount = excluded.amount,
		    updated_at = now(),
		    version = player_currencies.version + 1
		returning player_id, currency_type, amount, updated_at, version`,
		playerID, currencyType, amount,
	).Scan(&balance.PlayerID, &balance.CurrencyType, &balance.Amount, &balance.UpdatedAt, &balance.Version)
	if err != nil {
		return currencymodel.Balance{}, fmt.Errorf("upsert player %d currency %d: %w", playerID, currencyType, err)
	}

	return balance, nil
}

// insertLedger writes one currency audit entry.
func insertLedger(ctx context.Context, executor postgres.Executor, entry currencymodel.LedgerEntry) error {
	_, err := executor.Exec(ctx, `
		insert into currency_ledger_entries (
		    player_id, currency_type, delta, balance_after, reason, actor_kind, actor_id
		) values ($1, $2, $3, $4, $5, $6, $7)`,
		entry.PlayerID,
		entry.CurrencyType,
		entry.Delta,
		entry.BalanceAfter,
		entry.Reason,
		entry.ActorKind,
		entry.ActorID,
	)
	if err != nil {
		return fmt.Errorf("insert player %d currency %d ledger entry: %w", entry.PlayerID, entry.CurrencyType, err)
	}

	return nil
}
