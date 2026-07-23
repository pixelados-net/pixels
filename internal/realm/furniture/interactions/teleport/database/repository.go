// Package database persists furniture teleport pairs in PostgreSQL.
package database

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	teleportpair "github.com/niflaot/pixels/internal/realm/furniture/interactions/teleport/pair"
	"github.com/niflaot/pixels/pkg/postgres"
)

const (
	// findByItemSQL reads the relationship containing one item.
	findByItemSQL = `select item_one_id, item_two_id from furniture_item_teleport_pairs where item_one_id = $1 or item_two_id = $1`
	// deleteByItemSQL removes every relationship containing one item.
	deleteByItemSQL = `delete from furniture_item_teleport_pairs where item_one_id = $1 or item_two_id = $1`
	// insertPairSQL stores one canonical relationship.
	insertPairSQL = `insert into furniture_item_teleport_pairs (item_one_id, item_two_id) values ($1, $2)`
)

// Repository stores teleport pair relationships.
type Repository struct {
	// executor runs ordinary and scoped statements.
	executor postgres.Executor
	// within runs atomic replacement work.
	within func(context.Context, func(context.Context) error) error
}

// New creates a teleport pair repository.
func New(pool *postgres.Pool) *Repository {
	return &Repository{
		executor: pool,
		within: func(ctx context.Context, work func(context.Context) error) error {
			if _, ok := postgres.ScopedExecutor(ctx); ok {
				return work(ctx)
			}

			return postgres.WithinScope(ctx, pool, work)
		},
	}
}

// newRepository creates a repository around a test executor.
func newRepository(executor postgres.Executor) *Repository {
	return &Repository{executor: executor, within: func(ctx context.Context, work func(context.Context) error) error {
		return work(ctx)
	}}
}

// FindByItem finds a pair containing an item.
func (repository *Repository) FindByItem(ctx context.Context, itemID int64) (teleportpair.Pair, bool, error) {
	var paired teleportpair.Pair
	err := postgres.ExecutorFor(ctx, repository.executor).QueryRow(ctx, findByItemSQL, itemID).Scan(&paired.ItemOneID, &paired.ItemTwoID)
	if errors.Is(err, pgx.ErrNoRows) {
		return teleportpair.Pair{}, false, nil
	}
	if err != nil {
		return teleportpair.Pair{}, false, fmt.Errorf("find teleport pair for item %d: %w", itemID, err)
	}

	return paired, true, nil
}

// Replace atomically removes prior relationships and stores a pair.
func (repository *Repository) Replace(ctx context.Context, paired teleportpair.Pair) error {
	return repository.within(ctx, func(txCtx context.Context) error {
		executor := postgres.ExecutorFor(txCtx, repository.executor)
		if _, err := executor.Exec(txCtx, `delete from furniture_item_teleport_pairs where item_one_id in ($1, $2) or item_two_id in ($1, $2)`, paired.ItemOneID, paired.ItemTwoID); err != nil {
			return fmt.Errorf("replace existing teleport pairs: %w", err)
		}
		if _, err := executor.Exec(txCtx, insertPairSQL, paired.ItemOneID, paired.ItemTwoID); err != nil {
			return fmt.Errorf("store teleport pair: %w", err)
		}

		return nil
	})
}

// DeleteByItem removes the pair containing an item.
func (repository *Repository) DeleteByItem(ctx context.Context, itemID int64) (bool, error) {
	tag, err := postgres.ExecutorFor(ctx, repository.executor).Exec(ctx, deleteByItemSQL, itemID)
	if err != nil {
		return false, fmt.Errorf("delete teleport pair for item %d: %w", itemID, err)
	}

	return tag.RowsAffected() > 0, nil
}
