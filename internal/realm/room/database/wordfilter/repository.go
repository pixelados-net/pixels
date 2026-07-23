package repository

import (
	"context"
	"fmt"

	"github.com/niflaot/pixels/pkg/postgres"
)

const (
	// listSQL lists room filter words.
	listSQL = `select word from room_word_filters where room_id=$1 order by word`
	// addSQL inserts one room filter word.
	addSQL = `insert into room_word_filters (room_id, word) values ($1, $2) on conflict do nothing`
	// removeSQL removes one room filter word.
	removeSQL = `delete from room_word_filters where room_id=$1 and word=$2`
)

// Repository persists room word filters in PostgreSQL.
type Repository struct {
	// executor runs PostgreSQL queries.
	executor postgres.Executor
}

// New creates a room word filter repository.
func New(executor postgres.Executor) *Repository {
	return &Repository{executor: executor}
}

// List lists normalized words for a room.
func (repository *Repository) List(ctx context.Context, roomID int64) ([]string, error) {
	rows, err := repository.executor.Query(ctx, listSQL, roomID)
	if err != nil {
		return nil, fmt.Errorf("list room word filters: %w", err)
	}
	defer rows.Close()
	words := make([]string, 0, 8)
	for rows.Next() {
		var word string
		if err = rows.Scan(&word); err != nil {
			return nil, fmt.Errorf("scan room word filter: %w", err)
		}
		words = append(words, word)
	}

	return words, rows.Err()
}

// Add inserts a normalized room word.
func (repository *Repository) Add(ctx context.Context, roomID int64, word string) error {
	_, err := repository.executor.Exec(ctx, addSQL, roomID, word)
	if err != nil {
		return fmt.Errorf("add room word filter: %w", err)
	}

	return nil
}

// Remove deletes a normalized room word.
func (repository *Repository) Remove(ctx context.Context, roomID int64, word string) error {
	_, err := repository.executor.Exec(ctx, removeSQL, roomID, word)
	if err != nil {
		return fmt.Errorf("remove room word filter: %w", err)
	}

	return nil
}
