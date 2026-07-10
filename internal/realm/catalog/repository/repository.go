package repository

import (
	"context"

	"github.com/niflaot/pixels/pkg/postgres"
)

// transactionRunner runs catalog work in one transaction.
type transactionRunner func(context.Context, func(context.Context) error) error

// Repository reads and writes catalog records.
type Repository struct {
	// executor runs PostgreSQL queries.
	executor postgres.Executor

	// withinTx runs atomic catalog work.
	withinTx transactionRunner
}

// New creates a catalog repository.
func New(pool *postgres.Pool) *Repository {
	return &Repository{
		executor: pool,
		withinTx: func(ctx context.Context, work func(context.Context) error) error {
			return postgres.WithinScope(ctx, pool, work)
		},
	}
}

// newRepository creates a repository around a test executor.
func newRepository(executor postgres.Executor) *Repository {
	return &Repository{executor: executor, withinTx: func(ctx context.Context, work func(context.Context) error) error {
		return work(ctx)
	}}
}

// executorFor returns the active transaction or repository executor.
func (repository *Repository) executorFor(ctx context.Context) postgres.Executor {
	return postgres.ExecutorFor(ctx, repository.executor)
}

// WithinTransaction runs catalog purchase work atomically.
func (repository *Repository) WithinTransaction(ctx context.Context, work func(context.Context) error) error {
	if _, ok := postgres.ScopedExecutor(ctx); ok {
		return work(ctx)
	}

	return repository.withinTx(ctx, work)
}
