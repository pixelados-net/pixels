// Package database implements PostgreSQL pet persistence.
package database

import (
	"context"

	petpolicy "github.com/niflaot/pixels/internal/realm/pet/policy"
	"github.com/niflaot/pixels/pkg/postgres"
)

// Repository reads and writes pet aggregates.
type Repository struct {
	// config stores durable stat materialization policy.
	config petpolicy.Config
	// pool starts shared transaction scopes.
	pool *postgres.Pool
	// executor runs PostgreSQL operations.
	executor postgres.Executor
}

// New creates a pet repository.
func New(config petpolicy.Config, pool *postgres.Pool) *Repository {
	return &Repository{config: config.Normalize(), pool: pool, executor: pool}
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
