package repository

import (
	"context"

	"github.com/niflaot/pixels/internal/realm/room/control/rights"

	"github.com/niflaot/pixels/pkg/postgres"
)

// Repository persists room rights records.
type Repository struct {
	// executor runs PostgreSQL statements.
	executor postgres.Executor
	// withinTx runs shared transaction scopes.
	withinTx func(context.Context, func(context.Context) error) error
}

// New creates a room rights repository.
func New(pool *postgres.Pool) *Repository {
	return &Repository{executor: pool, withinTx: func(ctx context.Context, work func(context.Context) error) error {
		return postgres.WithinScope(ctx, pool, work)
	}}
}

// WithinTransaction runs work in one transaction.
func (repository *Repository) WithinTransaction(ctx context.Context, work rights.TransactionWork) error {
	if _, found := postgres.ScopedExecutor(ctx); found {
		return work(ctx)
	}

	return repository.withinTx(ctx, work)
}

// executorFor returns the scoped transaction or default executor.
func (repository *Repository) executorFor(ctx context.Context) postgres.Executor {
	return postgres.ExecutorFor(ctx, repository.executor)
}
