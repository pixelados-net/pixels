package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
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
)

// CreatePlayerParams contains player creation data.
type CreatePlayerParams struct {
	// Username is the unique visible player name.
	Username string
}

// CreatePlayer creates a player identity record.
func (repository *Repository) CreatePlayer(ctx context.Context, params CreatePlayerParams) (playermodel.Player, error) {
	player, err := scanPlayer(repository.executor.QueryRow(ctx, createPlayerSQL, params.Username))
	if err != nil {
		return playermodel.Player{}, fmt.Errorf("create player: %w", err)
	}

	return player, nil
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
	player, err := scanPlayer(repository.executor.QueryRow(ctx, query, argument))
	if errors.Is(err, pgx.ErrNoRows) {
		return playermodel.Player{}, false, nil
	}

	if err != nil {
		return playermodel.Player{}, false, err
	}

	return player, true, nil
}
