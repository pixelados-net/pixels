package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	roomlayout "github.com/niflaot/pixels/internal/realm/room/world/layout"
	"github.com/niflaot/pixels/pkg/postgres"
)

const (
	// findCustomByRoomIDSQL reads one active room-owned layout.
	findCustomByRoomIDSQL = `
select custom.room_id, rooms.model_name, custom.heightmap, custom.door_x, custom.door_y,
       custom.door_direction, custom.wall_thickness, custom.floor_thickness,
       custom.wall_height, custom.updated_at
from room_custom_layouts custom
join rooms on rooms.id = custom.room_id and rooms.deleted_at is null
where custom.room_id = $1`

	// upsertCustomSQL atomically saves geometry and synchronizes room visualization fields.
	upsertCustomSQL = `
with saved as (
    insert into room_custom_layouts
        (room_id, heightmap, door_x, door_y, door_direction, wall_thickness, floor_thickness, wall_height)
    values ($1, $2, $3, $4, $5, $6, $7, $8)
    on conflict (room_id) do update set
        heightmap = excluded.heightmap, door_x = excluded.door_x, door_y = excluded.door_y,
        door_direction = excluded.door_direction, wall_thickness = excluded.wall_thickness,
        floor_thickness = excluded.floor_thickness, wall_height = excluded.wall_height, updated_at = now()
    returning room_id, heightmap, door_x, door_y, door_direction, wall_thickness,
              floor_thickness, wall_height, updated_at
), updated_room as (
    update rooms set wall_thickness = $6, floor_thickness = $7,
                     updated_at = now(), version = version + 1
    where id = $1 and deleted_at is null
    returning model_name
)
select saved.room_id, updated_room.model_name, saved.heightmap, saved.door_x, saved.door_y,
       saved.door_direction, saved.wall_thickness, saved.floor_thickness,
       saved.wall_height, saved.updated_at
from saved cross join updated_room`
)

// FindCustomByRoomID finds a room's custom layout.
func (repository *Repository) FindCustomByRoomID(ctx context.Context, roomID int64) (roomlayout.Layout, bool, error) {
	roomLayout, err := scanCustomLayout(repository.executorFor(ctx).QueryRow(ctx, findCustomByRoomIDSQL, roomID))
	if errors.Is(err, pgx.ErrNoRows) {
		return roomlayout.Layout{}, false, nil
	}
	if err != nil {
		return roomlayout.Layout{}, false, fmt.Errorf("find custom room layout: %w", err)
	}

	return roomLayout, true, nil
}

// UpsertCustom creates or replaces a room's custom layout.
func (repository *Repository) UpsertCustom(ctx context.Context, params roomlayout.CustomSaveParams) (roomlayout.Layout, error) {
	row := repository.executorFor(ctx).QueryRow(ctx, upsertCustomSQL,
		params.RoomID, params.Heightmap, params.DoorX, params.DoorY, params.DoorDirection,
		params.WallThickness, params.FloorThickness, params.WallHeight,
	)
	roomLayout, err := scanCustomLayout(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return roomlayout.Layout{}, roomlayout.ErrLayoutNotFound
	}
	if err != nil {
		return roomlayout.Layout{}, fmt.Errorf("save custom room layout: %w", err)
	}

	return roomLayout, nil
}

// WithinTransaction runs custom layout work in one shared transaction.
func (repository *Repository) WithinTransaction(ctx context.Context, work roomlayout.TransactionWork) error {
	if repository.pool == nil {
		return work(ctx)
	}

	return postgres.WithinScope(ctx, repository.pool, work)
}

// customStoreAssertion verifies Repository implements custom layout persistence.
var customStoreAssertion roomlayout.CustomStore = (*Repository)(nil)
