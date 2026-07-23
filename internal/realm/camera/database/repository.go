// Package database implements PostgreSQL camera persistence.
package database

import (
	"context"

	camerarecord "github.com/niflaot/pixels/internal/realm/camera/record"
	"github.com/niflaot/pixels/pkg/postgres"
)

// Repository reads and writes camera aggregates.
type Repository struct {
	// pool starts shared transaction scopes.
	pool *postgres.Pool
	// executor runs PostgreSQL operations.
	executor postgres.Executor
}

// New creates a camera repository.
func New(pool *postgres.Pool) *Repository { return &Repository{pool: pool, executor: pool} }

// executorFor returns a transaction-scoped executor when present.
func (repository *Repository) executorFor(ctx context.Context) postgres.Executor {
	return postgres.ExecutorFor(ctx, repository.executor)
}

// WithinTransaction runs work in one shared PostgreSQL transaction.
func (repository *Repository) WithinTransaction(ctx context.Context, work func(context.Context) error) error {
	if _, found := postgres.ScopedExecutor(ctx); found {
		return work(ctx)
	}
	return postgres.WithinScope(ctx, repository.pool, work)
}

// storeAssertion verifies that Repository implements camera persistence.
var storeAssertion camerarecord.Store = (*Repository)(nil)
