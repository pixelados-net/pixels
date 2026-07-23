package repository

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/niflaot/pixels/pkg/postgres"
)

// Executor describes PostgreSQL operations used by permission persistence.
type Executor interface {
	// Exec executes a statement.
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	// Query executes a statement returning rows.
	Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error)
	// QueryRow executes a statement returning one row.
	QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row
}

// Repository implements permission persistence.
type Repository struct {
	// executor runs PostgreSQL statements.
	executor Executor
}

// New creates a permission repository.
func New(executor Executor) *Repository {
	return &Repository{executor: executor}
}

// NewFromPool creates a permission repository from the shared PostgreSQL pool.
func NewFromPool(pool *postgres.Pool) *Repository {
	return New(pool)
}
