package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	gamecenterrecord "github.com/niflaot/pixels/internal/realm/gamecenter/record"
)

// CreateGame inserts one external game registration.
func (repository *Repository) CreateGame(ctx context.Context, game gamecenterrecord.Game) (gamecenterrecord.Game, error) {
	err := repository.executorFor(ctx).QueryRow(ctx, `insert into game_center_games(name,bg_color,text_color,asset_url,support_url,launch_url,launch_kind,enabled) values($1,$2,$3,$4,$5,$6,$7,$8) returning id,version`, game.Name, game.BackgroundColor, game.TextColor, game.AssetURL, game.SupportURL, game.LaunchURL, game.LaunchKind, game.Enabled).Scan(&game.ID, &game.Version)
	return game, err
}

// UpdateGame replaces one registration with optimistic concurrency control.
func (repository *Repository) UpdateGame(ctx context.Context, game gamecenterrecord.Game) (gamecenterrecord.Game, bool, error) {
	err := repository.executorFor(ctx).QueryRow(ctx, `update game_center_games set name=$3,bg_color=$4,text_color=$5,asset_url=$6,support_url=$7,launch_url=$8,launch_kind=$9,enabled=$10,version=version+1,updated_at=now() where id=$1 and version=$2 returning version`, game.ID, game.Version, game.Name, game.BackgroundColor, game.TextColor, game.AssetURL, game.SupportURL, game.LaunchURL, game.LaunchKind, game.Enabled).Scan(&game.Version)
	if errors.Is(err, pgx.ErrNoRows) {
		return gamecenterrecord.Game{}, false, nil
	}
	return game, err == nil, err
}

// DisableGame idempotently disables one registration.
func (repository *Repository) DisableGame(ctx context.Context, id int32) (bool, error) {
	result, err := repository.executorFor(ctx).Exec(ctx, `update game_center_games set enabled=false,version=version+1,updated_at=now() where id=$1 and enabled`, id)
	return err == nil && result.RowsAffected() > 0, err
}
