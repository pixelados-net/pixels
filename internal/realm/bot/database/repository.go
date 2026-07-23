// Package database implements PostgreSQL bot persistence.
package database

import (
	"context"

	"github.com/niflaot/pixels/pkg/postgres"
)

// Repository reads and writes bot records.
type Repository struct {
	// pool starts shared transaction scopes.
	pool *postgres.Pool
	// executor runs PostgreSQL operations.
	executor postgres.Executor
}

// New creates a bot repository.
func New(pool *postgres.Pool) *Repository {
	return &Repository{pool: pool, executor: pool}
}

// executorFor returns a transaction-scoped executor when present.
func (repository *Repository) executorFor(ctx context.Context) postgres.Executor {
	return postgres.ExecutorFor(ctx, repository.executor)
}

// WithinTransaction runs work in the active transaction scope.
func (repository *Repository) WithinTransaction(ctx context.Context, work func(context.Context) error) error {
	if _, found := postgres.ScopedExecutor(ctx); found {
		return work(ctx)
	}
	return postgres.WithinScope(ctx, repository.pool, work)
}
