// Package database implements PostgreSQL Game Center persistence.
package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	gamecenterrecord "github.com/niflaot/pixels/internal/realm/gamecenter/record"
	"github.com/niflaot/pixels/pkg/postgres"
)

// Repository persists games and weekly scores.
type Repository struct {
	// pool starts shared transaction scopes.
	pool *postgres.Pool
	// executor runs PostgreSQL operations.
	executor postgres.Executor
}

// New creates a Game Center repository.
func New(pool *postgres.Pool) *Repository { return &Repository{pool: pool, executor: pool} }

// executorFor returns a transaction-scoped executor when present.
func (repository *Repository) executorFor(ctx context.Context) postgres.Executor {
	return postgres.ExecutorFor(ctx, repository.executor)
}

// ListGames returns games in stable id order.
func (repository *Repository) ListGames(ctx context.Context, enabledOnly bool) ([]gamecenterrecord.Game, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `select id, name, bg_color, text_color, asset_url, support_url, launch_url, launch_kind, enabled, version from game_center_games where not $1 or enabled order by id`, enabledOnly)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	games := make([]gamecenterrecord.Game, 0)
	for rows.Next() {
		var game gamecenterrecord.Game
		if err = rows.Scan(&game.ID, &game.Name, &game.BackgroundColor, &game.TextColor, &game.AssetURL, &game.SupportURL, &game.LaunchURL, &game.LaunchKind, &game.Enabled, &game.Version); err != nil {
			return nil, err
		}
		games = append(games, game)
	}
	return games, rows.Err()
}

// FindGame returns one game by id.
func (repository *Repository) FindGame(ctx context.Context, id int32) (gamecenterrecord.Game, bool, error) {
	var game gamecenterrecord.Game
	err := repository.executorFor(ctx).QueryRow(ctx, `select id, name, bg_color, text_color, asset_url, support_url, launch_url, launch_kind, enabled, version from game_center_games where id = $1`, id).Scan(&game.ID, &game.Name, &game.BackgroundColor, &game.TextColor, &game.AssetURL, &game.SupportURL, &game.LaunchURL, &game.LaunchKind, &game.Enabled, &game.Version)
	if errors.Is(err, pgx.ErrNoRows) {
		return gamecenterrecord.Game{}, false, nil
	}
	return game, err == nil, err
}

// UpsertScore keeps one player's best weekly score.
func (repository *Repository) UpsertScore(ctx context.Context, gameID int32, playerID int64, year int32, week int32, score int64) error {
	_, err := repository.executorFor(ctx).Exec(ctx, `insert into game_center_scores (game_id, player_id, year, week, score) values ($1,$2,$3,$4,$5) on conflict (game_id, player_id, year, week) do update set score = greatest(game_center_scores.score, excluded.score), updated_at = now()`, gameID, playerID, year, week, score)
	return err
}

var storeAssertion gamecenterrecord.Store = (*Repository)(nil)
