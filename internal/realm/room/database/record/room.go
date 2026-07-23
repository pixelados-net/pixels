package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	roomservice "github.com/niflaot/pixels/internal/realm/room/record/service"
	"github.com/niflaot/pixels/pkg/postgres"
)

const (
	// roomColumns contains the shared room select list.
	roomColumns = `id, owner_player_id, owner_name, name, description, model_name, door_mode, password_hash, max_users, score, category_id, trade_mode, roller_speed, allow_walkthrough, allow_pets, allow_pets_eat, hide_walls, hide_wired, wall_thickness, floor_thickness, chat_mode, chat_weight, chat_speed, chat_distance, chat_protection, moderation_mute, moderation_kick, moderation_ban, staff_picked, public_room, is_bundle_template, floor_paint, wallpaper, landscape, created_at, updated_at, deleted_at, version`

	// createRoomSQL inserts a room record.
	createRoomSQL = `
insert into rooms (owner_player_id, owner_name, name, description, model_name, door_mode, max_users, category_id, trade_mode)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
returning ` + roomColumns

	// findRoomByIDSQL reads one active room by id.
	findRoomByIDSQL = `select ` + roomColumns + ` from rooms where id = $1 and deleted_at is null`

	// listRoomsByIDsSQL reads active rooms in caller-provided identifier order.
	listRoomsByIDsSQL = `select ` + roomColumns + ` from rooms where id=any($1) and deleted_at is null and not is_bundle_template order by array_position($1::bigint[],id)`

	// listRoomsByOwnerSQL reads active rooms by owner.
	listRoomsByOwnerSQL = `select ` + roomColumns + ` from rooms where owner_player_id = $1 and deleted_at is null and not is_bundle_template order by id desc`

	// listRoomsByOwnerIDsSQL reads active rooms for several owners.
	listRoomsByOwnerIDsSQL = `select ` + roomColumns + ` from rooms where owner_player_id=any($1) and deleted_at is null and not is_bundle_template order by updated_at desc,id desc`

	// listPopularRoomsSQL reads active rooms by score and recency.
	listPopularRoomsSQL = `select ` + roomColumns + ` from rooms where deleted_at is null and not is_bundle_template and door_mode <> 3 order by score desc, updated_at desc limit $1`

	// listHighestScoreRoomsSQL reads active rooms by score.
	listHighestScoreRoomsSQL = `select ` + roomColumns + ` from rooms where deleted_at is null and not is_bundle_template and door_mode <> 3 order by score desc, id asc limit $1`

	// listOfficialRoomsSQL reads active staff-picked official rooms.
	listOfficialRoomsSQL = `select ` + roomColumns + ` from rooms where deleted_at is null and not is_bundle_template and (staff_picked or public_room) and door_mode <> 3 order by staff_picked desc,updated_at desc,id asc limit $1`

	// searchRoomsSQL reads active rooms matching public navigator text.
	searchRoomsSQL = `select ` + roomColumns + ` from rooms where deleted_at is null and not is_bundle_template and door_mode <> 3 and (name ilike $1 or owner_name ilike $1 or description ilike $1 or exists (select 1 from room_tags where room_tags.room_id = rooms.id and room_tags.tag ilike $1)) order by score desc, updated_at desc limit $2`

	// softDeleteRoomSQL soft deletes one active room.
	softDeleteRoomSQL = `update rooms set deleted_at = now(), updated_at = now(), version = version + 1 where id = $1 and deleted_at is null`

	// updateRoomSQL replaces editable settings using optimistic locking.
	updateRoomSQL = `
update rooms set name=$3, description=$4, door_mode=$5, password_hash=$6, max_users=$7,
category_id=$8, trade_mode=$9, roller_speed=$10, allow_walkthrough=$11, allow_pets=$12, allow_pets_eat=$13,
hide_walls=$14, hide_wired=$15, wall_thickness=$16, floor_thickness=$17, chat_mode=$18, chat_weight=$19,
chat_speed=$20, chat_distance=$21, chat_protection=$22, moderation_mute=$23,
moderation_kick=$24, moderation_ban=$25, staff_picked=$26, updated_at=now(), version=version+1
where id=$1 and version=$2 and deleted_at is null returning ` + roomColumns
)

// CreateRoom creates a room record.
func (repository *Repository) CreateRoom(ctx context.Context, params roomservice.CreateRecordParams) (roommodel.Room, error) {
	room, err := scanRoom(repository.executorFor(ctx).QueryRow(ctx, createRoomSQL, params.OwnerPlayerID, params.OwnerName, params.Name, params.Description, params.ModelName, params.DoorMode, params.MaxUsers, params.CategoryID, params.TradeMode))
	if err != nil {
		return roommodel.Room{}, fmt.Errorf("create room: %w", err)
	}

	return room, nil
}

// FindRoomByID finds an active room by id.
func (repository *Repository) FindRoomByID(ctx context.Context, id int64) (roommodel.Room, bool, error) {
	return repository.findRoom(ctx, findRoomByIDSQL, id)
}

// ListRoomsByIDs lists active rooms in caller-provided identifier order.
func (repository *Repository) ListRoomsByIDs(ctx context.Context, ids []int64) ([]roommodel.Room, error) {
	return repository.listRooms(ctx, listRoomsByIDsSQL, ids)
}

// ListRoomsByOwner lists active rooms owned by a player.
func (repository *Repository) ListRoomsByOwner(ctx context.Context, ownerPlayerID int64) ([]roommodel.Room, error) {
	return repository.listRooms(ctx, listRoomsByOwnerSQL, ownerPlayerID)
}

// ListRoomsByOwnerIDs lists active rooms for several owners.
func (repository *Repository) ListRoomsByOwnerIDs(ctx context.Context, ownerPlayerIDs []int64) ([]roommodel.Room, error) {
	return repository.listRooms(ctx, listRoomsByOwnerIDsSQL, ownerPlayerIDs)
}

// ListPopularRooms lists active rooms by occupancy-facing popularity fields.
func (repository *Repository) ListPopularRooms(ctx context.Context, limit int) ([]roommodel.Room, error) {
	return repository.listRooms(ctx, listPopularRoomsSQL, limit)
}

// ListHighestScoreRooms lists active rooms by score.
func (repository *Repository) ListHighestScoreRooms(ctx context.Context, limit int) ([]roommodel.Room, error) {
	return repository.listRooms(ctx, listHighestScoreRoomsSQL, limit)
}

// ListOfficialRooms lists active staff-picked official rooms.
func (repository *Repository) ListOfficialRooms(ctx context.Context, limit int) ([]roommodel.Room, error) {
	return repository.listRooms(ctx, listOfficialRoomsSQL, limit)
}

// SearchRooms searches active rooms by public navigator text.
func (repository *Repository) SearchRooms(ctx context.Context, query string, limit int) ([]roommodel.Room, error) {
	return repository.listRooms2(ctx, searchRoomsSQL, "%"+query+"%", limit)
}

// SoftDeleteRoom soft deletes a room record.
func (repository *Repository) SoftDeleteRoom(ctx context.Context, id int64) (bool, error) {
	tag, err := repository.executorFor(ctx).Exec(ctx, softDeleteRoomSQL, id)
	if err != nil {
		return false, fmt.Errorf("soft delete room: %w", err)
	}

	return tag.RowsAffected() > 0, nil
}

// UpdateRoom updates room settings and tags atomically with optimistic locking.
func (repository *Repository) UpdateRoom(ctx context.Context, params roomservice.UpdateRecordParams, tags []string) (roommodel.Room, bool, error) {
	var updated roommodel.Room
	var found bool
	err := repository.withinTx(ctx, func(txCtx context.Context, executor postgres.Executor) error {
		room := params.Room
		var err error
		updated, err = scanRoom(executor.QueryRow(txCtx, updateRoomSQL,
			room.ID, params.ExpectedVersion, room.Name, room.Description, room.DoorMode,
			room.PasswordHash, room.MaxUsers, room.CategoryID, room.TradeMode, room.RollerSpeed,
			room.AllowWalkthrough, room.AllowPets, room.AllowPetsEat, room.HideWalls,
			room.HideWired, room.WallThickness, room.FloorThickness, room.ChatMode, room.ChatWeight,
			room.ChatSpeed, room.ChatDistance, room.ChatProtection, room.ModerationMute,
			room.ModerationKick, room.ModerationBan, room.StaffPicked))
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("update room settings: %w", err)
		}
		found = true
		if _, err = executor.Exec(txCtx, deleteRoomTagsSQL, room.ID); err != nil {
			return fmt.Errorf("delete room tags: %w", err)
		}
		for _, tag := range tags {
			if _, err = executor.Exec(txCtx, insertRoomTagSQL, room.ID, tag); err != nil {
				return fmt.Errorf("insert room tag: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return roommodel.Room{}, false, err
	}

	return updated, found, nil
}

// findRoom finds one room with a query.
func (repository *Repository) findRoom(ctx context.Context, query string, argument any) (roommodel.Room, bool, error) {
	room, err := scanRoom(repository.executorFor(ctx).QueryRow(ctx, query, argument))
	if errors.Is(err, pgx.ErrNoRows) {
		return roommodel.Room{}, false, nil
	}

	if err != nil {
		return roommodel.Room{}, false, err
	}

	return room, true, nil
}

// listRooms lists rooms with a query.
func (repository *Repository) listRooms(ctx context.Context, query string, argument any) ([]roommodel.Room, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, query, argument)
	if err != nil {
		return nil, fmt.Errorf("list rooms: %w", err)
	}
	defer rows.Close()

	return scanRooms(rows)
}

// listRooms2 lists rooms with two query arguments.
func (repository *Repository) listRooms2(ctx context.Context, query string, first any, second any) ([]roommodel.Room, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, query, first, second)
	if err != nil {
		return nil, fmt.Errorf("list rooms: %w", err)
	}
	defer rows.Close()

	return scanRooms(rows)
}
