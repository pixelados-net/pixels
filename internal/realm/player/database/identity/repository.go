// Package identity implements PostgreSQL atomic player renames.
package identity

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	playeridentity "github.com/niflaot/pixels/internal/realm/player/identity"
	"github.com/niflaot/pixels/pkg/postgres"
)

// Repository commits player identity replacements.
type Repository struct {
	// pool owns transaction scopes.
	pool *postgres.Pool
}

// New creates an identity repository.
func New(pool *postgres.Pool) *Repository { return &Repository{pool: pool} }

// Rename commits one one-shot username replacement and audit row.
func (repository *Repository) Rename(ctx context.Context, playerID int64, username string) (result playeridentity.RenameResult, err error) {
	err = postgres.WithinScope(ctx, repository.pool, func(txCtx context.Context) error {
		executor := postgres.ExecutorFor(txCtx, repository.pool)
		var allowed bool
		lockErr := executor.QueryRow(txCtx, `select p.username,pp.allow_name_change from players p join player_profiles pp on pp.player_id=p.id where p.id=$1 and p.deleted_at is null for update of p,pp`, playerID).Scan(&result.OldUsername, &allowed)
		if errors.Is(lockErr, pgx.ErrNoRows) || lockErr == nil && !allowed {
			return playeridentity.ErrRenameDisabled
		}
		if lockErr != nil {
			return lockErr
		}
		if _, updateErr := executor.Exec(txCtx, `update players set username=$2,updated_at=now(),version=version+1 where id=$1`, playerID, username); updateErr != nil {
			return updateErr
		}
		if _, updateErr := executor.Exec(txCtx, `update player_profiles set allow_name_change=false,updated_at=now(),version=version+1 where player_id=$1`, playerID); updateErr != nil {
			return updateErr
		}
		if _, updateErr := executor.Exec(txCtx, `update rooms set owner_name=$2,updated_at=now(),version=version+1 where owner_player_id=$1 and deleted_at is null`, playerID, username); updateErr != nil {
			return updateErr
		}
		_, insertErr := executor.Exec(txCtx, `insert into player_name_changes(player_id,old_username,new_username,actor_player_id,reason,source) values($1,$2,$3,$1,'self-service','client')`, playerID, result.OldUsername, username)
		return insertErr
	})
	if err != nil {
		var postgresError *pgconn.PgError
		if errors.As(err, &postgresError) && postgresError.Code == "23505" {
			return playeridentity.RenameResult{}, playeridentity.ErrUsernameTaken
		}
		return playeridentity.RenameResult{}, fmt.Errorf("rename player: %w", err)
	}
	result.NewUsername = username
	return result, nil
}
