package repository

import "github.com/niflaot/pixels/pkg/postgres"

// Repository persists bubble thresholds in PostgreSQL.
type Repository struct {
	// pool executes PostgreSQL operations.
	pool *postgres.Pool
}

// New creates a bubble threshold repository.
func New(pool *postgres.Pool) *Repository {
	return &Repository{pool: pool}
}
