package repository

import (
	"context"

	"github.com/niflaot/pixels/pkg/postgres"
)

// Repository persists append-only room audit history.
type Repository struct {
	// executor runs PostgreSQL statements.
	executor postgres.Executor
}

// New creates a room audit repository.
func New(pool *postgres.Pool) *Repository {
	return &Repository{executor: pool}
}

// executorFor returns the scoped transaction or default executor.
func (repository *Repository) executorFor(ctx context.Context) postgres.Executor {
	return postgres.ExecutorFor(ctx, repository.executor)
}
