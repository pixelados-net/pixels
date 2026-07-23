package decoration

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	roomdecor "github.com/niflaot/pixels/internal/realm/room/decoration"
	"github.com/niflaot/pixels/pkg/postgres"
)

// dimmerMutation changes one room dimmer within a transaction.
type dimmerMutation func(context.Context, postgres.Executor) error

// changeDimmer validates room ownership, changes presets, and refreshes item state.
func (repository *Repository) changeDimmer(ctx context.Context, roomID int64, playerID int64, mutate dimmerMutation) (roomdecor.DimmerState, bool, error) {
	var state roomdecor.DimmerState
	changed := false
	err := postgres.WithinScope(ctx, repository.pool, func(txCtx context.Context) error {
		executor := postgres.ExecutorFor(txCtx, repository.pool)
		var ownerID int64
		if err := executor.QueryRow(txCtx, `select owner_player_id from rooms where id=$1 and deleted_at is null for update`, roomID).Scan(&ownerID); err != nil {
			return err
		}
		if ownerID != playerID {
			return nil
		}
		if err := initializePresets(txCtx, executor, roomID); err != nil {
			return err
		}
		if err := mutate(txCtx, executor); err != nil {
			return err
		}
		var found bool
		var err error
		state, found, err = repository.loadDimmer(txCtx, executor, roomID)
		if err != nil || !found {
			return err
		}
		result, err := executor.Exec(txCtx, `update furniture_items set extra_data=$2,updated_at=now(),version=version+1 where id=$1 and deleted_at is null`, state.ItemID, state.ExtraData)
		if err != nil {
			return err
		}
		changed = result.RowsAffected() == 1

		return nil
	})

	return state, changed, err
}

// loadDimmer reads one placed dimmer and ordered presets.
func (repository *Repository) loadDimmer(ctx context.Context, executor postgres.Executor, roomID int64) (roomdecor.DimmerState, bool, error) {
	var state roomdecor.DimmerState
	err := executor.QueryRow(ctx, `select fi.id from furniture_items fi join furniture_definitions fd on fd.id=fi.definition_id where fi.room_id=$1 and fi.deleted_at is null and fd.interaction_type='dimmer' order by fi.id limit 1`, roomID).Scan(&state.ItemID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return state, false, nil
		}
		return state, false, err
	}
	if err := initializePresets(ctx, executor, roomID); err != nil {
		return state, false, err
	}
	rows, err := executor.Query(ctx, `select preset_id,background_only,color,brightness,selected,enabled from room_dimmer_presets where room_id=$1 order by preset_id`, roomID)
	if err != nil {
		return state, false, err
	}
	defer rows.Close()
	for rows.Next() {
		var preset roomdecor.Preset
		if err = rows.Scan(&preset.ID, &preset.BackgroundOnly, &preset.Color, &preset.Brightness, &preset.Selected, &preset.Enabled); err != nil {
			return state, false, err
		}
		preset.Color = strings.TrimSpace(preset.Color)
		state.Presets = append(state.Presets, preset)
		if preset.Selected {
			mode := 1
			if preset.BackgroundOnly {
				mode = 2
			}
			enabled := 1
			if preset.Enabled {
				enabled = 2
			}
			state.ExtraData = fmt.Sprintf("%d,%d,%d,%s,%d", enabled, preset.ID, mode, preset.Color, preset.Brightness)
		}
	}

	return state, true, rows.Err()
}

// initializePresets creates the three deterministic default slots.
func initializePresets(ctx context.Context, executor postgres.Executor, roomID int64) error {
	_, err := executor.Exec(ctx, `insert into room_dimmer_presets(room_id,preset_id,selected) values($1,1,true),($1,2,false),($1,3,false) on conflict do nothing`, roomID)

	return err
}

// storeAssertion verifies the repository decoration boundary.
var storeAssertion roomdecor.Store = (*Repository)(nil)
