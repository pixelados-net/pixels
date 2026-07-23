// Package wired persists room WIRED configuration in PostgreSQL.
package wired

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	roomwired "github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/pkg/postgres"
)

const nodeSelect = `select fi.id,fi.room_id,fd.interaction_type,fd.sprite_id,fi.x,fi.y,coalesce(ws.int_params,'[]'::jsonb),coalesce(ws.string_param,''),coalesce(ws.selection_mode,0),coalesce(ws.delay_pulses,0),coalesce(ws.version,0) from furniture_items fi join furniture_definitions fd on fd.id=fi.definition_id left join room_wired_settings ws on ws.item_id=fi.id where fi.room_id=$1 and fi.deleted_at is null`

// Repository implements the WIRED domain store.
type Repository struct {
	// pool executes PostgreSQL operations.
	pool *postgres.Pool
}

// New creates a WIRED PostgreSQL repository.
func New(pool *postgres.Pool) *Repository { return &Repository{pool: pool} }

// LoadRoom loads configured WIRED nodes in one room.
func (repository *Repository) LoadRoom(ctx context.Context, roomID int64) ([]roomwired.Config, error) {
	rows, err := repository.pool.Query(ctx, nodeSelect+` and ws.item_id is not null order by fi.id`, roomID)
	if err != nil {
		return nil, fmt.Errorf("load room WIRED: %w", err)
	}
	defer rows.Close()
	configs := make([]roomwired.Config, 0)
	for rows.Next() {
		config, scanErr := scanConfig(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		config.Targets, scanErr = repository.loadTargets(ctx, repository.pool, config.ItemID)
		if scanErr != nil {
			return nil, scanErr
		}
		configs = append(configs, config)
	}
	return configs, rows.Err()
}

// Find loads one WIRED node and returns default settings when not configured.
func (repository *Repository) Find(ctx context.Context, roomID int64, itemID int64) (roomwired.Config, bool, error) {
	config, err := scanConfig(repository.pool.QueryRow(ctx, nodeSelect+` and fi.id=$2`, roomID, itemID))
	if errors.Is(err, pgx.ErrNoRows) {
		return roomwired.Config{}, false, nil
	}
	if err != nil {
		return roomwired.Config{}, false, err
	}
	config.Targets, err = repository.loadTargets(ctx, repository.pool, itemID)
	return config, true, err
}

// Save atomically replaces one configuration using optimistic versioning.
func (repository *Repository) Save(ctx context.Context, config roomwired.Config, expectedVersion int64) (roomwired.Config, error) {
	err := repository.within(ctx, func(txCtx context.Context) error {
		executor := postgres.ExecutorFor(txCtx, repository.pool)
		if err := lockItem(txCtx, executor, config.RoomID, config.ItemID); err != nil {
			return err
		}
		if err := validateTargets(txCtx, executor, config.RoomID, config.Targets); err != nil {
			return err
		}
		encoded, err := json.Marshal(config.IntParams)
		if err != nil {
			return err
		}
		if expectedVersion == 0 {
			var version int64
			err = executor.QueryRow(txCtx, `insert into room_wired_settings(item_id,int_params,string_param,selection_mode,delay_pulses) values($1,$2,$3,$4,$5) on conflict(item_id) do nothing returning version`, config.ItemID, encoded, config.StringParam, config.SelectionMode, config.DelayPulses).Scan(&version)
			if errors.Is(err, pgx.ErrNoRows) {
				return roomwired.ErrConflict
			}
			config.Version = version
		} else {
			err = executor.QueryRow(txCtx, `update room_wired_settings set int_params=$3,string_param=$4,selection_mode=$5,delay_pulses=$6,version=version+1,updated_at=now() where item_id=$1 and version=$2 returning version`, config.ItemID, expectedVersion, encoded, config.StringParam, config.SelectionMode, config.DelayPulses).Scan(&config.Version)
			if errors.Is(err, pgx.ErrNoRows) {
				return roomwired.ErrConflict
			}
		}
		if err != nil {
			return err
		}
		return replaceTargets(txCtx, executor, config.ItemID, config.Targets)
	})
	return config, err
}

// SaveRewardConfig atomically replaces settings, targets, and reward definitions.
func (repository *Repository) SaveRewardConfig(ctx context.Context, config roomwired.Config, expectedVersion int64, rewards []roomwired.Reward) (roomwired.Config, error) {
	err := repository.within(ctx, func(txCtx context.Context) error {
		saved, err := repository.Save(txCtx, config, expectedVersion)
		if err != nil {
			return err
		}
		config = saved
		return repository.ReplaceRewards(txCtx, config.ItemID, rewards)
	})
	return config, err
}

// Capture replaces snapshots from current authoritative furniture placement.
func (repository *Repository) Capture(ctx context.Context, roomID int64, itemID int64) ([]roomwired.Target, error) {
	var targets []roomwired.Target
	err := postgres.WithinScope(ctx, repository.pool, func(txCtx context.Context) error {
		executor := postgres.ExecutorFor(txCtx, repository.pool)
		if err := lockItem(txCtx, executor, roomID, itemID); err != nil {
			return err
		}
		rows, err := executor.Query(txCtx, `update room_wired_selected_items selected set snapshot_state=target.extra_data,snapshot_x=target.x,snapshot_y=target.y,snapshot_z=target.z,snapshot_rotation=target.rotation from furniture_items target join furniture_definitions definition on definition.id=target.definition_id where selected.wired_item_id=$1 and target.id=selected.selected_item_id and target.room_id=$2 and target.deleted_at is null returning selected.selected_item_id,definition.sprite_id,selected.snapshot_state,selected.snapshot_x,selected.snapshot_y,selected.snapshot_z,selected.snapshot_rotation,selected.ordinal`, itemID, roomID)
		if err != nil {
			return err
		}
		defer rows.Close()
		targets, err = scanTargets(rows)
		return err
	})
	return targets, err
}

// CleanupItem removes configuration owned by, and target references to, one picked-up item.
func (repository *Repository) CleanupItem(ctx context.Context, itemID int64) error {
	return repository.within(ctx, func(txCtx context.Context) error {
		executor := postgres.ExecutorFor(txCtx, repository.pool)
		if _, err := executor.Exec(txCtx, `delete from room_wired_selected_items where selected_item_id=$1`, itemID); err != nil {
			return fmt.Errorf("delete picked WIRED target: %w", err)
		}
		if _, err := executor.Exec(txCtx, `delete from room_wired_settings where item_id=$1`, itemID); err != nil {
			return fmt.Errorf("delete picked WIRED settings: %w", err)
		}
		return nil
	})
}
