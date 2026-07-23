// Package database implements subscription persistence with PostgreSQL.
package database

import (
	"context"

	"github.com/niflaot/pixels/internal/realm/subscription/record"
	"github.com/niflaot/pixels/pkg/postgres"
)

// Repository stores subscription records.
type Repository struct {
	// executor runs PostgreSQL queries.
	executor postgres.Executor
	// pool owns top-level transactions.
	pool *postgres.Pool
}

// New creates a subscription repository.
func New(pool *postgres.Pool) *Repository {
	return &Repository{executor: pool, pool: pool}
}

// executorFor returns the active scoped transaction.
func (repository *Repository) executorFor(ctx context.Context) postgres.Executor {
	return postgres.ExecutorFor(ctx, repository.executor)
}

// WithinTransaction runs work atomically.
func (repository *Repository) WithinTransaction(ctx context.Context, work func(context.Context) error) error {
	if _, found := postgres.ScopedExecutor(ctx); found {
		return work(ctx)
	}
	return postgres.WithinScope(ctx, repository.pool, work)
}

// storeAssertion verifies Repository implements Store.
var storeAssertion record.Store = (*Repository)(nil)
