// Package decoration contains PostgreSQL room decorator persistence.
package decoration

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	roomdecor "github.com/niflaot/pixels/internal/realm/room/decoration"
	"github.com/niflaot/pixels/pkg/postgres"
)

const (
	// placePostItSQL atomically places one owned post-it below the per-room limit.
	placePostItSQL = `with room_lock as (select pg_advisory_xact_lock($3)) update furniture_items fi set room_id=$3,x=null,y=null,z=null,rotation=0,wall_position=$4,extra_data=$5,updated_at=now(),version=fi.version+1 from furniture_definitions fd,room_lock where fi.id=$1 and fi.owner_player_id=$2 and fi.definition_id=fd.id and fd.interaction_type='postit' and fi.room_id is null and not fi.marketplace_reserved and fi.deleted_at is null and (select count(*) from furniture_items placed join furniture_definitions pfd on pfd.id=placed.definition_id where placed.room_id=$3 and placed.deleted_at is null and pfd.interaction_type='postit') < 200`
)

// Repository persists room decoration in focused transactions.
type Repository struct {
	// pool starts atomic decorator operations.
	pool *postgres.Pool
}

// New creates a room decoration repository.
func New(pool *postgres.Pool) *Repository {
	return &Repository{pool: pool}
}

// ConsumeSurface atomically validates and consumes a room-effect item.
func (repository *Repository) ConsumeSurface(ctx context.Context, itemID int64, playerID int64, roomID int64, surface roomdecor.Surface, value string) (bool, error) {
	changed := false
	err := postgres.WithinScope(ctx, repository.pool, func(txCtx context.Context) error {
		executor := postgres.ExecutorFor(txCtx, repository.pool)
		var name string
		var extraData string
		err := executor.QueryRow(txCtx, `select fd.name,fi.extra_data from furniture_items fi join furniture_definitions fd on fd.id=fi.definition_id where fi.id=$1 and fi.owner_player_id=$2 and fi.room_id is null and not fi.marketplace_reserved and fi.deleted_at is null for update`, itemID, playerID).Scan(&name, &extraData)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		if name != string(surface) || extraData != value {
			return nil
		}
		column := map[roomdecor.Surface]string{roomdecor.SurfaceFloor: "floor_paint", roomdecor.SurfaceWallpaper: "wallpaper", roomdecor.SurfaceLandscape: "landscape"}[surface]
		result, err := executor.Exec(txCtx, `update rooms set `+column+`=$2,updated_at=now(),version=version+1 where id=$1 and deleted_at is null`, roomID, value)
		if err != nil || result.RowsAffected() != 1 {
			return err
		}
		result, err = executor.Exec(txCtx, `update furniture_items set deleted_at=now(),updated_at=now(),version=version+1 where id=$1 and deleted_at is null`, itemID)
		if err != nil {
			return err
		}
		changed = result.RowsAffected() == 1

		return nil
	})

	return changed, err
}

// PlacePostIt atomically places one owned inventory post-it on a wall.
func (repository *Repository) PlacePostIt(ctx context.Context, itemID int64, playerID int64, roomID int64, wallPosition string, initialData string) (bool, error) {
	result, err := repository.pool.Exec(ctx, placePostItSQL, itemID, playerID, roomID, wallPosition, initialData)
	if err != nil {
		return false, fmt.Errorf("place post-it: %w", err)
	}

	return result.RowsAffected() == 1, nil
}

// LoadDimmer loads or initializes the room's three dimmer presets.
func (repository *Repository) LoadDimmer(ctx context.Context, roomID int64) (roomdecor.DimmerState, bool, error) {
	return repository.loadDimmer(ctx, postgres.ExecutorFor(ctx, repository.pool), roomID)
}

// SaveDimmer atomically saves and optionally selects one preset.
func (repository *Repository) SaveDimmer(ctx context.Context, roomID int64, playerID int64, preset roomdecor.Preset, apply bool) (roomdecor.DimmerState, bool, error) {
	return repository.changeDimmer(ctx, roomID, playerID, func(txCtx context.Context, executor postgres.Executor) error {
		if apply {
			if _, err := executor.Exec(txCtx, `update room_dimmer_presets set selected=false where room_id=$1`, roomID); err != nil {
				return err
			}
		}
		_, err := executor.Exec(txCtx, `insert into room_dimmer_presets(room_id,preset_id,background_only,color,brightness,selected,enabled) values($1,$2,$3,$4,$5,$6,$6) on conflict(room_id,preset_id) do update set background_only=excluded.background_only,color=excluded.color,brightness=excluded.brightness,selected=case when $6 then true else room_dimmer_presets.selected end,enabled=case when $6 then true else room_dimmer_presets.enabled end,updated_at=now()`, roomID, preset.ID, preset.BackgroundOnly, preset.Color, preset.Brightness, apply)

		return err
	})
}

// ToggleDimmer atomically toggles the selected preset.
func (repository *Repository) ToggleDimmer(ctx context.Context, roomID int64, playerID int64) (roomdecor.DimmerState, bool, error) {
	return repository.changeDimmer(ctx, roomID, playerID, func(txCtx context.Context, executor postgres.Executor) error {
		_, err := executor.Exec(txCtx, `update room_dimmer_presets set enabled=not enabled,updated_at=now() where room_id=$1 and selected`, roomID)

		return err
	})
}
