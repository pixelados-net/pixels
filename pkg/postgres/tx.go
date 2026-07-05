package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Executor describes common PostgreSQL query behavior.
type Executor interface {
	// Exec executes SQL without returning rows.
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)

	// Query executes SQL returning multiple rows.
	Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error)

	// QueryRow executes SQL returning one row.
	QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row
}

// TxFunc runs database work inside one transaction.
type TxFunc func(ctx context.Context, tx pgx.Tx) error

// WithinTx runs work inside a transaction.
func WithinTx(ctx context.Context, pool *Pool, fn TxFunc) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin postgres transaction: %w", err)
	}

	return finishTx(ctx, tx, fn(ctx, tx))
}

// transaction describes transaction completion behavior.
type transaction interface {
	// Commit commits a transaction.
	Commit(ctx context.Context) error

	// Rollback rolls back a transaction.
	Rollback(ctx context.Context) error
}

// rollbacker describes transaction rollback behavior.
type rollbacker interface {
	// Rollback rolls back a transaction.
	Rollback(ctx context.Context) error
}

// finishTx commits or rolls back transaction work.
func finishTx(ctx context.Context, tx transaction, cause error) error {
	if cause != nil {
		return rollback(ctx, tx, cause)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit postgres transaction: %w", err)
	}

	return nil
}

// rollback rolls back a failed transaction.
func rollback(ctx context.Context, tx rollbacker, cause error) error {
	if err := tx.Rollback(ctx); err != nil {
		return fmt.Errorf("rollback postgres transaction after %w: %w", cause, err)
	}

	return cause
}
