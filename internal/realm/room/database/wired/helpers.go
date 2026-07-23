package wired

import (
	"context"
	"encoding/json"
	"errors"
	"sort"

	"github.com/jackc/pgx/v5"
	roomwired "github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/pkg/postgres"
)

// rowScanner scans one PostgreSQL row.
type rowScanner interface {
	// Scan assigns row columns to destinations.
	Scan(...any) error
}

// scanConfig scans one node plus optional persisted settings.
func scanConfig(row rowScanner) (roomwired.Config, error) {
	var config roomwired.Config
	var encoded []byte
	err := row.Scan(&config.ItemID, &config.RoomID, &config.Interaction, &config.SpriteID, &config.X, &config.Y, &encoded, &config.StringParam, &config.SelectionMode, &config.DelayPulses, &config.Version)
	if err != nil {
		return roomwired.Config{}, err
	}
	if err := json.Unmarshal(encoded, &config.IntParams); err != nil {
		return roomwired.Config{}, err
	}
	return config, nil
}

// lockItem locks and verifies one placed WIRED item.
func lockItem(ctx context.Context, executor postgres.Executor, roomID int64, itemID int64) error {
	var found int64
	err := executor.QueryRow(ctx, `select id from furniture_items where id=$1 and room_id=$2 and deleted_at is null for update`, itemID, roomID).Scan(&found)
	if errors.Is(err, pgx.ErrNoRows) {
		return roomwired.ErrItemMissing
	}
	return err
}

// validateTargets verifies that every selected item remains in the same room.
func validateTargets(ctx context.Context, executor postgres.Executor, roomID int64, targets []roomwired.Target) error {
	seen := make(map[int64]struct{}, len(targets))
	for _, target := range targets {
		if target.ItemID <= 0 {
			return roomwired.ErrTargetMissing
		}
		if _, exists := seen[target.ItemID]; exists {
			return roomwired.ErrTargetMissing
		}
		seen[target.ItemID] = struct{}{}
		var found int64
		err := executor.QueryRow(ctx, `select id from furniture_items where id=$1 and room_id=$2 and deleted_at is null`, target.ItemID, roomID).Scan(&found)
		if errors.Is(err, pgx.ErrNoRows) {
			return roomwired.ErrTargetMissing
		}
		if err != nil {
			return err
		}
	}
	return nil
}

// replaceTargets replaces ordered targets inside the caller transaction.
func replaceTargets(ctx context.Context, executor postgres.Executor, itemID int64, targets []roomwired.Target) error {
	if _, err := executor.Exec(ctx, `delete from room_wired_selected_items where wired_item_id=$1`, itemID); err != nil {
		return err
	}
	for ordinal, target := range targets {
		var state any
		var x any
		var y any
		var z any
		var rotation any
		if target.Snapshot.Present {
			state, x, y, z, rotation = target.Snapshot.State, target.Snapshot.X, target.Snapshot.Y, target.Snapshot.Z, target.Snapshot.Rotation
		}
		_, err := executor.Exec(ctx, `insert into room_wired_selected_items(wired_item_id,selected_item_id,ordinal,snapshot_state,snapshot_x,snapshot_y,snapshot_z,snapshot_rotation) values($1,$2,$3,$4,$5,$6,$7,$8)`, itemID, target.ItemID, ordinal, state, x, y, z, rotation)
		if err != nil {
			return err
		}
	}
	return nil
}

// loadTargets loads ordered selected items.
func (repository *Repository) loadTargets(ctx context.Context, executor postgres.Executor, itemID int64) ([]roomwired.Target, error) {
	rows, err := executor.Query(ctx, `select selected.selected_item_id,definition.sprite_id,selected.snapshot_state,selected.snapshot_x,selected.snapshot_y,selected.snapshot_z,selected.snapshot_rotation,selected.ordinal from room_wired_selected_items selected join furniture_items item on item.id=selected.selected_item_id join furniture_items wired on wired.id=selected.wired_item_id join furniture_definitions definition on definition.id=item.definition_id where selected.wired_item_id=$1 and item.room_id=wired.room_id and item.deleted_at is null order by selected.ordinal`, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanTargets(rows)
}

// scanTargets scans and sorts selected targets by persisted ordinal.
func scanTargets(rows pgx.Rows) ([]roomwired.Target, error) {
	type ordered struct {
		target  roomwired.Target
		ordinal int
	}
	values := make([]ordered, 0)
	for rows.Next() {
		var value ordered
		var state *string
		var x *int
		var y *int
		var z *float64
		var rotation *int
		if err := rows.Scan(&value.target.ItemID, &value.target.SpriteID, &state, &x, &y, &z, &rotation, &value.ordinal); err != nil {
			return nil, err
		}
		if state != nil && x != nil && y != nil && z != nil && rotation != nil {
			value.target.Snapshot = roomwired.Snapshot{State: *state, X: *x, Y: *y, Z: *z, Rotation: *rotation, Present: true}
		}
		values = append(values, value)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	sort.Slice(values, func(left int, right int) bool { return values[left].ordinal < values[right].ordinal })
	result := make([]roomwired.Target, len(values))
	for index := range values {
		result[index] = values[index].target
	}
	return result, nil
}
