package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	roombundle "github.com/niflaot/pixels/internal/realm/room/record/bundle"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	"github.com/niflaot/pixels/pkg/postgres"
)

const (
	// countRoomsByOwnerSQL counts ordinary active rooms.
	countRoomsByOwnerSQL = `select count(*) from rooms where owner_player_id=$1 and deleted_at is null and not is_bundle_template`
	// lockRoomOwnerSQL serializes ownership-limit checks.
	lockRoomOwnerSQL = `select pg_advisory_xact_lock(hashtextextended('room-owner:' || ($1::bigint)::text, 0))`
	// cloneBundleRoomSQL clones durable room settings and resets public ranking state.
	cloneBundleRoomSQL = `
insert into rooms
    (owner_player_id, owner_name, name, description, model_name, door_mode, password_hash,
     max_users, score, category_id, trade_mode, roller_speed, allow_walkthrough, allow_pets, allow_pets_eat,
     hide_walls, hide_wired, wall_thickness, floor_thickness, chat_mode, chat_weight, chat_speed,
     chat_distance, chat_protection, moderation_mute, moderation_kick, moderation_ban,
     staff_picked, public_room, is_bundle_template, floor_paint, wallpaper, landscape)
select $2, $3, name, description, model_name, door_mode, password_hash,
       max_users, 0, category_id, trade_mode, roller_speed, allow_walkthrough, allow_pets, allow_pets_eat,
       hide_walls, hide_wired, wall_thickness, floor_thickness, chat_mode, chat_weight, chat_speed,
       chat_distance, chat_protection, moderation_mute, moderation_kick, moderation_ban,
       false, false, false, floor_paint, wallpaper, landscape
from rooms where id=$1 and deleted_at is null and is_bundle_template
returning ` + roomColumns
	// recordBundlePurchaseSQL records template provenance.
	recordBundlePurchaseSQL = `insert into room_bundle_purchases (catalog_item_id,template_room_id,created_room_id,buyer_player_id,furniture_item_count,bot_count) values ($1,$2,$3,$4,$5,$6)`
	// setBundleTemplateSQL changes a room's template state.
	setBundleTemplateSQL = `update rooms set is_bundle_template=$2,updated_at=now(),version=version+1 where id=$1 and deleted_at is null returning ` + roomColumns
	// countActiveBundleReferencesSQL counts enabled active catalog references.
	countActiveBundleReferencesSQL = `select count(*) from catalog_items where room_bundle_template_room_id=$1 and enabled and deleted_at is null`
	// listBundleTemplateRoomsSQL lists active bundle templates.
	listBundleTemplateRoomsSQL = `select ` + roomColumns + ` from rooms where is_bundle_template and deleted_at is null order by id`
)

// WithinTransaction runs bundle work in the active PostgreSQL scope.
func (repository *Repository) WithinTransaction(ctx context.Context, work func(context.Context) error) error {
	if _, ok := postgres.ScopedExecutor(ctx); ok {
		return work(ctx)
	}
	return repository.withinTx(ctx, func(txCtx context.Context, _ postgres.Executor) error { return work(txCtx) })
}

// LockRoomOwner serializes room-limit checks for one owner.
func (repository *Repository) LockRoomOwner(ctx context.Context, ownerPlayerID int64) error {
	if _, err := repository.executorFor(ctx).Exec(ctx, lockRoomOwnerSQL, ownerPlayerID); err != nil {
		return fmt.Errorf("lock room owner %d: %w", ownerPlayerID, err)
	}
	return nil
}

// CountRoomsByOwner counts ordinary active rooms.
func (repository *Repository) CountRoomsByOwner(ctx context.Context, ownerPlayerID int64) (int, error) {
	var count int
	if err := repository.executorFor(ctx).QueryRow(ctx, countRoomsByOwnerSQL, ownerPlayerID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count rooms by owner: %w", err)
	}
	return count, nil
}

// CloneBundleRoom clones a marked template room for a buyer.
func (repository *Repository) CloneBundleRoom(ctx context.Context, templateRoomID int64, buyerPlayerID int64, buyerName string) (roommodel.Room, error) {
	room, err := scanRoom(repository.executorFor(ctx).QueryRow(ctx, cloneBundleRoomSQL, templateRoomID, buyerPlayerID, buyerName))
	if errors.Is(err, pgx.ErrNoRows) {
		return roommodel.Room{}, roombundle.ErrInvalidTemplate
	}
	if err != nil {
		return roommodel.Room{}, fmt.Errorf("clone bundle room: %w", err)
	}
	_, err = repository.executorFor(ctx).Exec(ctx, `insert into room_dimmer_presets(room_id,preset_id,background_only,color,brightness,selected,enabled) select $2,preset_id,background_only,color,brightness,selected,enabled from room_dimmer_presets where room_id=$1`, templateRoomID, room.ID)
	if err != nil {
		return roommodel.Room{}, fmt.Errorf("clone bundle dimmer presets: %w", err)
	}
	return room, nil
}

// RecordBundlePurchase records bundle provenance.
func (repository *Repository) RecordBundlePurchase(ctx context.Context, params roombundle.PurchaseRecord) error {
	_, err := repository.executorFor(ctx).Exec(ctx, recordBundlePurchaseSQL, params.CatalogItemID, params.TemplateRoomID, params.CreatedRoomID, params.BuyerPlayerID, params.FurnitureCount, params.BotCount)
	if err != nil {
		return fmt.Errorf("record room bundle purchase: %w", err)
	}
	return nil
}

// SetBundleTemplate changes a room's bundle source status.
func (repository *Repository) SetBundleTemplate(ctx context.Context, roomID int64, enabled bool) (roommodel.Room, bool, error) {
	room, err := scanRoom(repository.executorFor(ctx).QueryRow(ctx, setBundleTemplateSQL, roomID, enabled))
	if errors.Is(err, pgx.ErrNoRows) {
		return roommodel.Room{}, false, nil
	}
	return room, err == nil, err
}

// CountActiveBundleReferences counts enabled offers referencing a room.
func (repository *Repository) CountActiveBundleReferences(ctx context.Context, roomID int64) (int, error) {
	var count int
	if err := repository.executorFor(ctx).QueryRow(ctx, countActiveBundleReferencesSQL, roomID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count bundle template references: %w", err)
	}
	return count, nil
}

// ListBundleTemplateRooms lists marked active templates.
func (repository *Repository) ListBundleTemplateRooms(ctx context.Context) ([]roommodel.Room, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, listBundleTemplateRoomsSQL)
	if err != nil {
		return nil, fmt.Errorf("list room bundle templates: %w", err)
	}
	defer rows.Close()
	return scanRooms(rows)
}

// bundleStoreAssertion verifies Repository implements bundle persistence.
var bundleStoreAssertion roombundle.Store = (*Repository)(nil)
