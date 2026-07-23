package database

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5"
)

// TokenBalance reads a player's listing tokens.
func (repository *Repository) TokenBalance(ctx context.Context, playerID int64) (int32, error) {
	var amount int32
	err := repository.executor(ctx).QueryRow(ctx, `select amount from marketplace_tokens where player_id=$1`, playerID).Scan(&amount)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, nil
	}
	return amount, err
}

// AddTokens atomically adds a non-negative token package.
func (repository *Repository) AddTokens(ctx context.Context, playerID int64, amount int32) (int32, error) {
	var balance int32
	err := repository.executor(ctx).QueryRow(ctx, `insert into marketplace_tokens(player_id,amount) values($1,$2) on conflict(player_id) do update set amount=marketplace_tokens.amount+excluded.amount,updated_at=now() returning amount`, playerID, amount).Scan(&balance)
	return balance, err
}

// SpendToken atomically spends one token.
func (repository *Repository) SpendToken(ctx context.Context, playerID int64) (bool, error) {
	result, err := repository.executor(ctx).Exec(ctx, `update marketplace_tokens set amount=amount-1,updated_at=now() where player_id=$1 and amount>0`, playerID)
	return err == nil && result.RowsAffected() == 1, err
}
