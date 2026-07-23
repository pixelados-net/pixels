// Package repository contains PostgreSQL access for player records.
package repository

import (
	"context"

	"github.com/niflaot/pixels/pkg/postgres"
)

// Repository reads and writes player persistence records.
type Repository struct {
	// executor runs PostgreSQL queries.
	executor postgres.Executor

	// pool owns transaction boundaries when the executor is a PostgreSQL pool.
	pool *postgres.Pool
}

// New creates a player repository.
func New(executor postgres.Executor) *Repository {
	repository := &Repository{executor: executor}
	repository.pool, _ = executor.(*postgres.Pool)

	return repository
}

// WithinTransaction runs work in a shared PostgreSQL transaction.
func (repository *Repository) WithinTransaction(ctx context.Context, work func(context.Context) error) error {
	if repository.pool == nil {
		return work(ctx)
	}

	return postgres.WithinScope(ctx, repository.pool, work)
}
