// Package repository contains PostgreSQL access for room records.
package repository

import "github.com/niflaot/pixels/pkg/postgres"

// Repository reads and writes room persistence records.
type Repository struct {
	// executor runs PostgreSQL queries.
	executor postgres.Executor
}

// New creates a room repository.
func New(executor postgres.Executor) *Repository {
	return &Repository{executor: executor}
}
