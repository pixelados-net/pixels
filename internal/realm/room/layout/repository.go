package layout

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/niflaot/pixels/pkg/postgres"
)

const (
	// createLayoutSQL inserts a room layout record.
	createLayoutSQL = `
insert into room_layouts (name, tile_size, heightmap, door_x, door_y, door_z, door_direction, club_level, enabled)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
returning id, name, tile_size, heightmap, door_x, door_y, door_z, door_direction, club_level, enabled, created_at, updated_at, deleted_at, version`

	// updateLayoutSQL updates an active room layout record.
	updateLayoutSQL = `
update room_layouts
set name = $2, tile_size = $3, heightmap = $4, door_x = $5, door_y = $6, door_z = $7, door_direction = $8, club_level = $9, enabled = $10, updated_at = now(), version = version + 1
where id = $1 and deleted_at is null
returning id, name, tile_size, heightmap, door_x, door_y, door_z, door_direction, club_level, enabled, created_at, updated_at, deleted_at, version`

	// findLayoutByIDSQL reads one active room layout by id.
	findLayoutByIDSQL = `
select id, name, tile_size, heightmap, door_x, door_y, door_z, door_direction, club_level, enabled, created_at, updated_at, deleted_at, version
from room_layouts
where id = $1 and deleted_at is null`

	// findLayoutByNameSQL reads one active room layout by name.
	findLayoutByNameSQL = `
select id, name, tile_size, heightmap, door_x, door_y, door_z, door_direction, club_level, enabled, created_at, updated_at, deleted_at, version
from room_layouts
where name = $1 and deleted_at is null`

	// listLayoutsSQL reads active room layouts.
	listLayoutsSQL = `
select id, name, tile_size, heightmap, door_x, door_y, door_z, door_direction, club_level, enabled, created_at, updated_at, deleted_at, version
from room_layouts
where deleted_at is null
order by name asc`
)

// Repository reads and writes room layout records.
type Repository struct {
	// executor runs PostgreSQL queries.
	executor postgres.Executor
}

// Reader reads persistent room layouts.
type Reader interface {
	// FindByID finds an active room layout by id.
	FindByID(ctx context.Context, id int64) (Layout, bool, error)

	// FindByName finds an active room layout by normalized name.
	FindByName(ctx context.Context, name string) (Layout, bool, error)

	// List lists active room layouts.
	List(ctx context.Context) ([]Layout, error)
}

// Writer writes persistent room layouts.
type Writer interface {
	// Create creates a room layout record.
	Create(ctx context.Context, params CreateRecordParams) (Layout, error)

	// Update updates a room layout record.
	Update(ctx context.Context, params UpdateRecordParams) (Layout, bool, error)
}

// Store reads and writes persistent room layouts.
type Store interface {
	Reader
	Writer
}

// CreateRecordParams contains room layout creation data.
type CreateRecordParams struct {
	// Layout contains the room layout values.
	Layout Layout
}

// UpdateRecordParams contains room layout update data.
type UpdateRecordParams struct {
	// ID identifies the room layout record.
	ID int64

	// Layout contains the replacement room layout values.
	Layout Layout
}

// NewRepository creates a room layout repository.
func NewRepository(executor postgres.Executor) *Repository {
	return &Repository{executor: executor}
}

// Create creates a room layout record.
func (repository *Repository) Create(ctx context.Context, params CreateRecordParams) (Layout, error) {
	roomLayout, err := scanLayout(repository.executor.QueryRow(ctx, createLayoutSQL, layoutValues(params.Layout)...))
	if err != nil {
		return Layout{}, fmt.Errorf("create room layout: %w", err)
	}

	return roomLayout, nil
}

// Update updates a room layout record.
func (repository *Repository) Update(ctx context.Context, params UpdateRecordParams) (Layout, bool, error) {
	values := append([]any{params.ID}, layoutValues(params.Layout)...)
	roomLayout, err := scanLayout(repository.executor.QueryRow(ctx, updateLayoutSQL, values...))
	if errors.Is(err, pgx.ErrNoRows) {
		return Layout{}, false, nil
	}

	if err != nil {
		return Layout{}, false, fmt.Errorf("update room layout: %w", err)
	}

	return roomLayout, true, nil
}

// FindByID finds an active room layout by id.
func (repository *Repository) FindByID(ctx context.Context, id int64) (Layout, bool, error) {
	return repository.find(ctx, findLayoutByIDSQL, id)
}

// FindByName finds an active room layout by normalized name.
func (repository *Repository) FindByName(ctx context.Context, name string) (Layout, bool, error) {
	return repository.find(ctx, findLayoutByNameSQL, name)
}

// List lists active room layouts.
func (repository *Repository) List(ctx context.Context) ([]Layout, error) {
	rows, err := repository.executor.Query(ctx, listLayoutsSQL)
	if err != nil {
		return nil, fmt.Errorf("list room layouts: %w", err)
	}
	defer rows.Close()

	return scanLayouts(rows)
}

// find finds one room layout with a query.
func (repository *Repository) find(ctx context.Context, query string, argument any) (Layout, bool, error) {
	roomLayout, err := scanLayout(repository.executor.QueryRow(ctx, query, argument))
	if errors.Is(err, pgx.ErrNoRows) {
		return Layout{}, false, nil
	}

	if err != nil {
		return Layout{}, false, err
	}

	return roomLayout, true, nil
}

// layoutValues returns SQL arguments for editable room layout fields.
func layoutValues(roomLayout Layout) []any {
	return []any{
		roomLayout.Name,
		roomLayout.TileSize,
		roomLayout.Heightmap,
		roomLayout.DoorX,
		roomLayout.DoorY,
		roomLayout.DoorZ,
		roomLayout.DoorDirection,
		roomLayout.ClubLevel,
		roomLayout.Enabled,
	}
}

// storeAssertion verifies Repository implements Store.
var storeAssertion Store = (*Repository)(nil)
