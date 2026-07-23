// Package repository contains PostgreSQL access for navigator records.
package database

import (
	"github.com/niflaot/pixels/internal/realm/navigator/record"
	"github.com/niflaot/pixels/pkg/postgres"
)

// Repository reads and writes navigator persistence records.
type Repository struct {
	// executor runs PostgreSQL queries.
	executor postgres.Executor
}

// New creates a navigator repository.
func New(executor postgres.Executor) *Repository {
	return &Repository{executor: executor}
}

// storeAssertion verifies Repository implements Store.
var storeAssertion record.Store = (*Repository)(nil)
