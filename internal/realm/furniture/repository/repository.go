// Package repository contains PostgreSQL access for furniture records.
package repository

import "github.com/niflaot/pixels/pkg/postgres"

// Repository reads and writes furniture persistence records.
type Repository struct {
	// executor runs PostgreSQL queries.
	executor postgres.Executor
}

// New creates a furniture repository.
func New(executor postgres.Executor) *Repository {
	return &Repository{executor: executor}
}
