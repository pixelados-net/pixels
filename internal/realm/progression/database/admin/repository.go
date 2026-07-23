// Package admin implements PostgreSQL progression administration.
package admin

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	progressionrecord "github.com/niflaot/pixels/internal/realm/progression/record"
	"github.com/niflaot/pixels/pkg/postgres"
)

// Repository persists progression administration records.
type Repository struct {
	// pool starts shared transaction scopes.
	pool *postgres.Pool
	// executor runs PostgreSQL operations.
	executor postgres.Executor
}

// New creates one progression administration repository.
func New(pool *postgres.Pool) *Repository { return &Repository{pool: pool, executor: pool} }

// WithinTransaction runs work in one shared PostgreSQL transaction.
func (repository *Repository) WithinTransaction(ctx context.Context, work func(context.Context) error) error {
	if _, found := postgres.ScopedExecutor(ctx); found {
		return work(ctx)
	}
	return postgres.WithinScope(ctx, repository.pool, work)
}

// InsertAudit appends one administrative mutation record.
func (repository *Repository) InsertAudit(ctx context.Context, actorID int64, action string, entity string, reason string) error {
	_, err := repository.executorFor(ctx).Exec(ctx, `insert into progression_audit(actor_player_id,action,entity,reason) values(nullif($1,0),$2,$3,$4)`, actorID, action, entity, reason)
	return err
}

// executorFor returns a transaction-scoped executor when present.
func (repository *Repository) executorFor(ctx context.Context) postgres.Executor {
	return postgres.ExecutorFor(ctx, repository.executor)
}

// missingOrConflict distinguishes an absent row from an optimistic conflict.
func (repository *Repository) missingOrConflict(ctx context.Context, table string, column string, value any) error {
	var exists bool
	query := `select exists(select 1 from ` + table + ` where ` + column + `=$1)`
	if err := repository.executorFor(ctx).QueryRow(ctx, query, value).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return progressionrecord.ErrConflict
	}
	return progressionrecord.ErrNotFound
}

// noRows reports whether one query returned no row.
func noRows(err error) bool { return errors.Is(err, pgx.ErrNoRows) }

var storeAssertion progressionrecord.AdminStore = (*Repository)(nil)
