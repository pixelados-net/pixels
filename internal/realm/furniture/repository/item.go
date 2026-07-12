package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
)

const (
	// itemColumns contains the shared furniture item select list.
	itemColumns = `id, definition_id, owner_player_id, room_id, x, y, z::float8, rotation, wall_position, extra_data, metadata, created_at, updated_at, deleted_at, version`

	// findItemByIDSQL reads one active furniture item by id.
	findItemByIDSQL = `select ` + itemColumns + ` from furniture_items where id = $1 and deleted_at is null`

	// listInventoryItemsSQL reads active unplaced items owned by a player.
	listInventoryItemsSQL = `select ` + itemColumns + ` from furniture_items where owner_player_id = $1 and room_id is null and deleted_at is null order by id asc`

	// listRoomItemsSQL reads active items placed in a room.
	listRoomItemsSQL = `select ` + itemColumns + ` from furniture_items where room_id = $1 and deleted_at is null order by id asc`

	// createItemsSQL creates inventory items in one statement.
	createItemsSQL = `
insert into furniture_items (definition_id, owner_player_id, extra_data)
select $1, $2, $4 from generate_series(1, $3)
returning ` + itemColumns

	// placeItemSQL moves an owned inventory item into a room.
	placeItemSQL = `
update furniture_items
set room_id = $3, x = $4, y = $5, z = $6, rotation = $7, updated_at = now(), version = version + 1
where id = $1 and owner_player_id = $2 and room_id is null and deleted_at is null
returning ` + itemColumns

	// moveItemSQL repositions an item within its authorized room.
	moveItemSQL = `
update furniture_items
set x = $3, y = $4, z = $5, rotation = $6, updated_at = now(), version = version + 1
where id = $1 and room_id = $2 and deleted_at is null
returning ` + itemColumns

	// pickupItemSQL returns an owned, placed item to inventory.
	pickupItemSQL = `
update furniture_items
set room_id = null, x = null, y = null, z = null, updated_at = now(), version = version + 1
where id = $1 and owner_player_id = $2 and room_id is not null and deleted_at is null
returning ` + itemColumns

	// updateItemStateSQL changes state only when the runtime snapshot still matches persistence.
	updateItemStateSQL = `
update furniture_items
set extra_data = $4, updated_at = now(), version = version + 1
where id = $1 and room_id = $2 and extra_data = $3 and deleted_at is null
returning ` + itemColumns
)

// PlaceItemParams contains input for placing an owned inventory item into a room.
type PlaceItemParams struct {
	// ID identifies the furniture item.
	ID int64

	// OwnerPlayerID identifies the required current owner.
	OwnerPlayerID int64

	// RoomID identifies the destination room.
	RoomID int64

	// Placement stores the destination floor coordinates and rotation.
	Placement furnituremodel.Placement
}

// MoveItemParams contains input for repositioning an item within one room.
type MoveItemParams struct {
	// ID identifies the furniture item.
	ID int64

	// RoomID identifies the required current room.
	RoomID int64

	// Placement stores the destination floor coordinates and rotation.
	Placement furnituremodel.Placement
}

// PickupItemParams contains input for returning an owned, placed item to inventory.
type PickupItemParams struct {
	// ID identifies the furniture item.
	ID int64

	// OwnerPlayerID identifies the required current owner.
	OwnerPlayerID int64
}

// UpdateItemStateParams contains one guarded furniture state mutation.
type UpdateItemStateParams struct {
	// ID identifies the furniture item.
	ID int64

	// RoomID identifies the required current room.
	RoomID int64

	// Expected stores the state observed by the active room.
	Expected string

	// Next stores the state to persist.
	Next string
}

// CreateItems creates inventory items for one owner and definition.
func (repository *Repository) CreateItems(ctx context.Context, definitionID int64, ownerPlayerID int64, quantity int32, extraData string) ([]furnituremodel.Item, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, createItemsSQL, definitionID, ownerPlayerID, quantity, extraData)
	if err != nil {
		return nil, fmt.Errorf("create %d furniture items for player %d: %w", quantity, ownerPlayerID, err)
	}
	defer rows.Close()

	return scanItems(rows)
}

// FindItemByID finds an active furniture item by id.
func (repository *Repository) FindItemByID(ctx context.Context, id int64) (furnituremodel.Item, bool, error) {
	return repository.queryItem(ctx, findItemByIDSQL, id)
}

// ListInventoryItems lists active unplaced items owned by a player.
func (repository *Repository) ListInventoryItems(ctx context.Context, ownerPlayerID int64) ([]furnituremodel.Item, error) {
	return repository.listItems(ctx, listInventoryItemsSQL, ownerPlayerID)
}

// ListRoomItems lists active items placed in a room.
func (repository *Repository) ListRoomItems(ctx context.Context, roomID int64) ([]furnituremodel.Item, error) {
	return repository.listItems(ctx, listRoomItemsSQL, roomID)
}

// PlaceItem moves an owned inventory item into a room.
func (repository *Repository) PlaceItem(ctx context.Context, params PlaceItemParams) (furnituremodel.Item, bool, error) {
	return repository.queryItem(ctx, placeItemSQL, params.ID, params.OwnerPlayerID, params.RoomID, params.Placement.X, params.Placement.Y, params.Placement.Z, params.Placement.Rotation)
}

// MoveItem repositions an item guarded by its current room.
func (repository *Repository) MoveItem(ctx context.Context, params MoveItemParams) (furnituremodel.Item, bool, error) {
	return repository.queryItem(ctx, moveItemSQL, params.ID, params.RoomID, params.Placement.X, params.Placement.Y, params.Placement.Z, params.Placement.Rotation)
}

// PickupItem returns an owned, placed item to inventory.
func (repository *Repository) PickupItem(ctx context.Context, params PickupItemParams) (furnituremodel.Item, bool, error) {
	return repository.queryItem(ctx, pickupItemSQL, params.ID, params.OwnerPlayerID)
}

// UpdateItemState changes one placed item's state with compare-and-swap semantics.
func (repository *Repository) UpdateItemState(ctx context.Context, params UpdateItemStateParams) (furnituremodel.Item, bool, error) {
	return repository.queryItem(ctx, updateItemStateSQL, params.ID, params.RoomID, params.Expected, params.Next)
}

// queryItem runs one row-returning query and reports whether a row was found.
func (repository *Repository) queryItem(ctx context.Context, query string, arguments ...any) (furnituremodel.Item, bool, error) {
	item, err := scanItem(repository.executorFor(ctx).QueryRow(ctx, query, arguments...))
	if errors.Is(err, pgx.ErrNoRows) {
		return furnituremodel.Item{}, false, nil
	}
	if err != nil {
		return furnituremodel.Item{}, false, err
	}

	return item, true, nil
}

// listItems lists furniture items with a query.
func (repository *Repository) listItems(ctx context.Context, query string, argument any) ([]furnituremodel.Item, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, query, argument)
	if err != nil {
		return nil, fmt.Errorf("list furniture items: %w", err)
	}
	defer rows.Close()

	return scanItems(rows)
}
