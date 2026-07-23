package repository

import "github.com/niflaot/pixels/pkg/postgres"

// Repository persists global chat filter words in PostgreSQL.
type Repository struct {
	// pool executes PostgreSQL operations.
	pool *postgres.Pool
}

// New creates a global chat filter repository.
func New(pool *postgres.Pool) *Repository {
	return &Repository{pool: pool}
}
