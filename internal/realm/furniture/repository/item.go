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
	itemColumns = `id, definition_id, owner_player_id, room_id, x, y, z::float8, rotation, wall_position, extra_data, limited_edition_number, marketplace_reserved, gift_wrapped, gift_wrap_sprite_id, gift_wrap_box_id, gift_wrap_ribbon_id, gift_sender_player_id, gift_message, metadata, created_at, updated_at, deleted_at, version`

	// findItemByIDSQL reads one active furniture item by id.
	findItemByIDSQL = `select ` + itemColumns + ` from furniture_items where id = $1 and deleted_at is null`

	// listInventoryItemsSQL reads active unplaced items owned by a player.
	listInventoryItemsSQL = `select ` + itemColumns + ` from furniture_items where owner_player_id = $1 and room_id is null and not marketplace_reserved and deleted_at is null order by id asc`

	// listRoomItemsSQL reads active items placed in a room.
	listRoomItemsSQL = `select ` + itemColumns + ` from furniture_items where room_id = $1 and deleted_at is null order by id asc`

	// createItemsSQL creates inventory items in one statement.
	createItemsSQL = `
insert into furniture_items (definition_id, owner_player_id, extra_data, limited_edition_number)
select $1, $2, $4, $5 from generate_series(1, $3)
returning ` + itemColumns

	// createGiftItemsSQL creates wrapped inventory items in one statement.
	createGiftItemsSQL = `insert into furniture_items (definition_id,owner_player_id,extra_data,gift_wrapped,gift_wrap_sprite_id,gift_wrap_box_id,gift_wrap_ribbon_id,gift_sender_player_id,gift_message) select $1,$2,$4,true,$5,$6,$7,$8,$9 from generate_series(1,$3) returning ` + itemColumns

	// placeItemSQL moves an owned inventory item into a room.
	placeItemSQL = `
with room_lock as (select pg_advisory_xact_lock($3))
update furniture_items
set room_id = $3,
    x = case when $8::text = '' then $4::smallint else null::smallint end,
    y = case when $8::text = '' then $5::smallint else null::smallint end,
    z = case when $8::text = '' then $6::numeric(6,2) else null::numeric(6,2) end,
    rotation = case when $8::text = '' then $7::smallint else 0::smallint end,
    wall_position = nullif($8::text, ''), updated_at = now(), version = version + 1
from room_lock
where id = $1 and owner_player_id = $2 and room_id is null and deleted_at is null
  and ($9::text = '' or not exists (
      select 1 from furniture_items placed
      join furniture_definitions definition on definition.id = placed.definition_id
      where placed.room_id = $3 and placed.deleted_at is null and definition.interaction_type = $9::text
  ))
returning ` + itemColumns

	// moveItemSQL repositions an item within its authorized room.
	moveItemSQL = `
update furniture_items
set x = case when $7::text = '' then $3::smallint else null::smallint end,
    y = case when $7::text = '' then $4::smallint else null::smallint end,
    z = case when $7::text = '' then $5::numeric(6,2) else null::numeric(6,2) end,
    rotation = case when $7::text = '' then $6::smallint else 0::smallint end,
    wall_position = nullif($7::text, ''), updated_at = now(), version = version + 1
where id = $1 and room_id = $2 and deleted_at is null
returning ` + itemColumns

	// pickupItemSQL returns an owned, placed item to inventory.
	pickupItemSQL = `
update furniture_items
set room_id = null, x = null, y = null, z = null, wall_position = null, updated_at = now(), version = version + 1
where id = $1 and owner_player_id = $2 and room_id is not null and deleted_at is null
returning ` + itemColumns

	// updateItemStateSQL changes state only when the runtime snapshot still matches persistence.
	updateItemStateSQL = `
update furniture_items
set extra_data = $4, updated_at = now(), version = version + 1
where id = $1 and room_id = $2 and extra_data = $3 and deleted_at is null
returning ` + itemColumns

	// openGiftItemSQL marks one placed gift as opened.
	openGiftItemSQL = `
update furniture_items
set gift_wrapped = false,
    gift_wrap_sprite_id = null,
    gift_wrap_box_id = null,
    gift_wrap_ribbon_id = null,
    gift_sender_player_id = null,
    gift_message = null,
    updated_at = now(),
    version = version + 1
where id = $1 and owner_player_id = $2 and room_id = $3 and gift_wrapped = true and deleted_at is null
returning ` + itemColumns

	// reserveForMarketplaceSQL withdraws an inventory item while retaining its valid owner FK.
	reserveForMarketplaceSQL = `update furniture_items set marketplace_reserved=true,updated_at=now(),version=version+1 where id=$1 and owner_player_id=$2 and room_id is null and not marketplace_reserved and deleted_at is null`

	// releaseFromMarketplaceSQL returns a reserved item to the seller inventory.
	releaseFromMarketplaceSQL = `update furniture_items set marketplace_reserved=false,updated_at=now(),version=version+1 where id=$1 and owner_player_id=$2 and marketplace_reserved and deleted_at is null`

	// transferFromMarketplaceSQL delivers a reserved item to its buyer.
	transferFromMarketplaceSQL = `update furniture_items set owner_player_id=$3,marketplace_reserved=false,updated_at=now(),version=version+1 where id=$1 and owner_player_id=$2 and marketplace_reserved and deleted_at is null`

	// transferInventoryItemSQL transfers one available inventory item.
	transferInventoryItemSQL = `update furniture_items set owner_player_id=$3,updated_at=now(),version=version+1 where id=$1 and owner_player_id=$2 and room_id is null and not marketplace_reserved and deleted_at is null`

	// deleteInventoryItemSQL consumes one available inventory item.
	deleteInventoryItemSQL = `update furniture_items set deleted_at=now(),updated_at=now(),version=version+1 where id=$1 and owner_player_id=$2 and room_id is null and not marketplace_reserved and deleted_at is null`
)

// GiftItemParams contains wrapped item persistence input.
type GiftItemParams struct {
	// DefinitionID identifies the furniture definition.
	DefinitionID int64
	// OwnerPlayerID identifies the recipient.
	OwnerPlayerID int64
	// Quantity stores the instance count.
	Quantity int32
	// ExtraData stores initial furniture state.
	ExtraData string
	// SpriteID stores the selected wrapping furniture sprite.
	SpriteID int32
	// BoxID stores the wrapping box.
	BoxID int32
	// RibbonID stores the wrapping ribbon.
	RibbonID int32
	// SenderPlayerID identifies the sender when visible.
	SenderPlayerID *int64
	// Message stores the gift message.
	Message string
}

// CreateItems creates inventory items for one owner and definition.
func (repository *Repository) CreateItems(ctx context.Context, definitionID int64, ownerPlayerID int64, quantity int32, extraData string, limitedEditionNumber *int32) ([]furnituremodel.Item, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, createItemsSQL, definitionID, ownerPlayerID, quantity, extraData, limitedEditionNumber)
	if err != nil {
		return nil, fmt.Errorf("create %d furniture items for player %d: %w", quantity, ownerPlayerID, err)
	}
	defer rows.Close()

	return scanItems(rows)
}

// CreateGiftItems creates wrapped inventory items for one recipient.
func (repository *Repository) CreateGiftItems(ctx context.Context, params GiftItemParams) ([]furnituremodel.Item, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, createGiftItemsSQL, params.DefinitionID, params.OwnerPlayerID, params.Quantity, params.ExtraData, params.SpriteID, params.BoxID, params.RibbonID, params.SenderPlayerID, params.Message)
	if err != nil {
		return nil, fmt.Errorf("create wrapped furniture for player %d: %w", params.OwnerPlayerID, err)
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

// ReserveForMarketplace withdraws one owned inventory item into Marketplace limbo.
func (repository *Repository) ReserveForMarketplace(ctx context.Context, itemID int64, ownerPlayerID int64) (bool, error) {
	return repository.execGuard(ctx, reserveForMarketplaceSQL, itemID, ownerPlayerID)
}

// ReleaseFromMarketplace returns one reserved item to its seller inventory.
func (repository *Repository) ReleaseFromMarketplace(ctx context.Context, itemID int64, ownerPlayerID int64) (bool, error) {
	return repository.execGuard(ctx, releaseFromMarketplaceSQL, itemID, ownerPlayerID)
}

// TransferFromMarketplace delivers one reserved item to its buyer.
func (repository *Repository) TransferFromMarketplace(ctx context.Context, itemID int64, sellerPlayerID int64, buyerPlayerID int64) (bool, error) {
	return repository.execGuard(ctx, transferFromMarketplaceSQL, itemID, sellerPlayerID, buyerPlayerID)
}

// TransferInventoryItem transfers one unreserved inventory item between players.
func (repository *Repository) TransferInventoryItem(ctx context.Context, itemID int64, fromPlayerID int64, toPlayerID int64) (bool, error) {
	return repository.execGuard(ctx, transferInventoryItemSQL, itemID, fromPlayerID, toPlayerID)
}

// DeleteInventoryItem soft-deletes one unreserved owned inventory item.
func (repository *Repository) DeleteInventoryItem(ctx context.Context, itemID int64, ownerPlayerID int64) (bool, error) {
	return repository.execGuard(ctx, deleteInventoryItemSQL, itemID, ownerPlayerID)
}

// execGuard executes a guarded item mutation and reports whether it changed a row.
func (repository *Repository) execGuard(ctx context.Context, query string, arguments ...any) (bool, error) {
	result, err := repository.executorFor(ctx).Exec(ctx, query, arguments...)
	if err != nil {
		return false, err
	}
	return result.RowsAffected() == 1, nil
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
