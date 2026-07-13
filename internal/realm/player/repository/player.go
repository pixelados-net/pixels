package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
	"github.com/niflaot/pixels/pkg/postgres"
)

const (
	// createPlayerSQL inserts a player identity record.
	createPlayerSQL = `
insert into players (username)
values ($1)
returning id, username, created_at, updated_at, deleted_at, version, last_login_at, last_logout_at, last_seen_at, club_level, club_expires_at`

	// findPlayerByIDSQL reads one active player by id.
	findPlayerByIDSQL = `
select id, username, created_at, updated_at, deleted_at, version, last_login_at, last_logout_at, last_seen_at, club_level, club_expires_at
from players
where id = $1 and deleted_at is null`

	// findPlayerByUsernameSQL reads one active player by username.
	findPlayerByUsernameSQL = `
select id, username, created_at, updated_at, deleted_at, version, last_login_at, last_logout_at, last_seen_at, club_level, club_expires_at
from players
where lower(username) = lower($1) and deleted_at is null`

	// updateClubSQL updates the derived player club entitlement.
	updateClubSQL = `update players set club_level=$2, club_expires_at=$3, updated_at=now(), version=version+1 where id=$1 and deleted_at is null`

	// updatePlayerSQL replaces one active player username using optimistic locking.
	updatePlayerSQL = `
update players set username=$2, updated_at=now(), version=version+1
where id=$1 and version=$3 and deleted_at is null
returning id, username, created_at, updated_at, deleted_at, version, last_login_at, last_logout_at, last_seen_at, club_level, club_expires_at`

	// softDeletePlayerSQL marks one active player deleted using optimistic locking.
	softDeletePlayerSQL = `update players set deleted_at=now(), updated_at=now(), version=version+1 where id=$1 and version=$2 and deleted_at is null`
)

// CreatePlayerParams contains player creation data.
type CreatePlayerParams struct {
	// Username is the unique visible player name.
	Username string
}

// UpdatePlayerParams contains one complete player identity replacement.
type UpdatePlayerParams struct {
	// PlayerID identifies the player.
	PlayerID int64
	// Username replaces the visible player name.
	Username string
	// ExpectedVersion prevents lost concurrent updates.
	ExpectedVersion int64
}

// UpdateClub updates the derived player club entitlement.
func (repository *Repository) UpdateClub(ctx context.Context, playerID int64, club playermodel.Club) error {
	_, err := postgres.ExecutorFor(ctx, repository.executor).Exec(ctx, updateClubSQL, playerID, club.Level, club.ExpiresAt)
	return err
}

// CreatePlayer creates a player identity record.
func (repository *Repository) CreatePlayer(ctx context.Context, params CreatePlayerParams) (playermodel.Player, error) {
	player, err := scanPlayer(postgres.ExecutorFor(ctx, repository.executor).QueryRow(ctx, createPlayerSQL, params.Username))
	if err != nil {
		var postgresError *pgconn.PgError
		if errors.As(err, &postgresError) && postgresError.Code == "23505" {
			return playermodel.Player{}, ErrUsernameTaken
		}
		return playermodel.Player{}, fmt.Errorf("create player: %w", err)
	}

	return player, nil
}

// UpdatePlayer updates one active player identity with optimistic locking.
func (repository *Repository) UpdatePlayer(ctx context.Context, params UpdatePlayerParams) (playermodel.Player, bool, error) {
	player, err := scanPlayer(postgres.ExecutorFor(ctx, repository.executor).QueryRow(ctx, updatePlayerSQL, params.PlayerID, params.Username, params.ExpectedVersion))
	if errors.Is(err, pgx.ErrNoRows) {
		return playermodel.Player{}, false, nil
	}
	if err != nil {
		var postgresError *pgconn.PgError
		if errors.As(err, &postgresError) && postgresError.Code == "23505" {
			return playermodel.Player{}, false, ErrUsernameTaken
		}
		return playermodel.Player{}, false, fmt.Errorf("update player: %w", err)
	}

	return player, true, nil
}

// SoftDeletePlayer marks one active player deleted with optimistic locking.
func (repository *Repository) SoftDeletePlayer(ctx context.Context, playerID int64, expectedVersion int64) (bool, error) {
	result, err := postgres.ExecutorFor(ctx, repository.executor).Exec(ctx, softDeletePlayerSQL, playerID, expectedVersion)
	if err != nil {
		return false, fmt.Errorf("soft delete player: %w", err)
	}

	return result.RowsAffected() == 1, nil
}

// FindPlayerByID finds an active player by id.
func (repository *Repository) FindPlayerByID(ctx context.Context, id int64) (playermodel.Player, bool, error) {
	player, found, err := repository.findPlayer(ctx, findPlayerByIDSQL, id)
	if err != nil {
		return playermodel.Player{}, false, fmt.Errorf("find player by id: %w", err)
	}

	return player, found, nil
}

// FindPlayerByUsername finds an active player by username.
func (repository *Repository) FindPlayerByUsername(ctx context.Context, username string) (playermodel.Player, bool, error) {
	player, found, err := repository.findPlayer(ctx, findPlayerByUsernameSQL, username)
	if err != nil {
		return playermodel.Player{}, false, fmt.Errorf("find player by username: %w", err)
	}

	return player, found, nil
}

// findPlayer finds one player with a query.
func (repository *Repository) findPlayer(ctx context.Context, query string, argument any) (playermodel.Player, bool, error) {
	player, err := scanPlayer(postgres.ExecutorFor(ctx, repository.executor).QueryRow(ctx, query, argument))
	if errors.Is(err, pgx.ErrNoRows) {
		return playermodel.Player{}, false, nil
	}

	if err != nil {
		return playermodel.Player{}, false, err
	}

	return player, true, nil
}
