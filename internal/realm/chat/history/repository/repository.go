package repository

import "github.com/niflaot/pixels/pkg/postgres"

// Repository persists chat history in PostgreSQL partitions.
type Repository struct {
	// pool executes PostgreSQL operations.
	pool *postgres.Pool
}

// New creates a partitioned chat history repository.
func New(pool *postgres.Pool) *Repository { return &Repository{pool: pool} }
