// Package repository contains PostgreSQL access for room records.
package repository

import (
	"context"

	"github.com/niflaot/pixels/pkg/postgres"
)

// transactionRunner runs repository work atomically.
type transactionRunner func(context.Context, func(context.Context, postgres.Executor) error) error

// Repository reads and writes room persistence records.
type Repository struct {
	// executor runs PostgreSQL queries.
	executor postgres.Executor

	// withinTx runs atomic room mutations.
	withinTx transactionRunner
}

// New creates a room repository.
func New(executor postgres.Executor) *Repository {
	repository := &Repository{executor: executor}
	pool, ok := executor.(*postgres.Pool)
	if ok {
		repository.withinTx = func(ctx context.Context, work func(context.Context, postgres.Executor) error) error {
			return postgres.WithinScope(ctx, pool, func(txCtx context.Context) error {
				return work(txCtx, postgres.ExecutorFor(txCtx, executor))
			})
		}
	} else {
		repository.withinTx = func(ctx context.Context, work func(context.Context, postgres.Executor) error) error {
			return work(ctx, executor)
		}
	}

	return repository
}
