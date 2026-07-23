package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// List returns configured thresholds.
func (repository *Repository) List(ctx context.Context) ([]Unlock, error) {
	rows, err := repository.pool.Query(ctx, `select bubble_id, min_weight from chat_bubble_unlocks order by bubble_id`)
	if err != nil {
		return nil, fmt.Errorf("list chat bubble unlocks: %w", err)
	}
	defer rows.Close()
	items := make([]Unlock, 0)
	for rows.Next() {
		var item Unlock
		if err = rows.Scan(&item.BubbleID, &item.MinWeight); err != nil {
			return nil, fmt.Errorf("scan chat bubble unlock: %w", err)
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

// MinWeight returns one threshold and whether it exists.
func (repository *Repository) MinWeight(ctx context.Context, bubbleID int32) (int32, bool, error) {
	var weight int32
	err := repository.pool.QueryRow(ctx, `select min_weight from chat_bubble_unlocks where bubble_id=$1`, bubbleID).Scan(&weight)
	if errors.Is(err, pgx.ErrNoRows) {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, fmt.Errorf("find chat bubble %d unlock: %w", bubbleID, err)
	}

	return weight, true, nil
}

// Set creates or replaces one threshold.
func (repository *Repository) Set(ctx context.Context, bubbleID int32, minWeight int32) error {
	_, err := repository.pool.Exec(ctx, `insert into chat_bubble_unlocks (bubble_id, min_weight) values ($1,$2) on conflict (bubble_id) do update set min_weight=excluded.min_weight`, bubbleID, minWeight)
	if err != nil {
		return fmt.Errorf("set chat bubble %d unlock: %w", bubbleID, err)
	}

	return nil
}

// Delete removes one threshold.
func (repository *Repository) Delete(ctx context.Context, bubbleID int32) error {
	_, err := repository.pool.Exec(ctx, `delete from chat_bubble_unlocks where bubble_id=$1`, bubbleID)
	if err != nil {
		return fmt.Errorf("delete chat bubble %d unlock: %w", bubbleID, err)
	}

	return nil
}
