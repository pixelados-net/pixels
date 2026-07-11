package repository

import (
	"context"
	"fmt"
)

// List returns normalized filter words.
func (repository *Repository) List(ctx context.Context) ([]string, error) {
	rows, err := repository.pool.Query(ctx, `select word from chat_global_word_filters order by word`)
	if err != nil {
		return nil, fmt.Errorf("list global chat filter words: %w", err)
	}
	defer rows.Close()
	words := make([]string, 0)
	for rows.Next() {
		var word string
		if err = rows.Scan(&word); err != nil {
			return nil, fmt.Errorf("scan global chat filter word: %w", err)
		}
		words = append(words, word)
	}

	return words, rows.Err()
}

// Add creates a filter word when absent.
func (repository *Repository) Add(ctx context.Context, word string) error {
	_, err := repository.pool.Exec(ctx, `insert into chat_global_word_filters (word) values ($1) on conflict (word) do nothing`, word)
	if err != nil {
		return fmt.Errorf("add global chat filter word %q: %w", word, err)
	}

	return nil
}

// Remove deletes a filter word when present.
func (repository *Repository) Remove(ctx context.Context, word string) error {
	_, err := repository.pool.Exec(ctx, `delete from chat_global_word_filters where word=$1`, word)
	if err != nil {
		return fmt.Errorf("remove global chat filter word %q: %w", word, err)
	}

	return nil
}
