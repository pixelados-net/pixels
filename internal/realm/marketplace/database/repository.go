// Package database implements Marketplace persistence with PostgreSQL.
package database

import (
	"context"

	marketrecord "github.com/niflaot/pixels/internal/realm/marketplace/record"
	"github.com/niflaot/pixels/pkg/postgres"
)

// Repository stores Marketplace records.
type Repository struct {
	// pool owns database transactions.
	pool *postgres.Pool
}

// New creates a Marketplace repository.
func New(pool *postgres.Pool) *Repository { return &Repository{pool: pool} }

// executor returns the active scoped transaction or pool.
func (repository *Repository) executor(ctx context.Context) postgres.Executor {
	return postgres.ExecutorFor(ctx, repository.pool)
}

// WithinTransaction runs work atomically.
func (repository *Repository) WithinTransaction(ctx context.Context, work func(context.Context) error) error {
	if _, ok := postgres.ScopedExecutor(ctx); ok {
		return work(ctx)
	}
	return postgres.WithinScope(ctx, repository.pool, work)
}

var storeAssertion marketrecord.Store = (*Repository)(nil)
