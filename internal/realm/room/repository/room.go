package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	roommodel "github.com/niflaot/pixels/internal/realm/room/model"
)

const (
	// roomColumns contains the shared room select list.
	roomColumns = `id, owner_player_id, owner_name, name, description, model_name, door_mode, password_hash, max_users, score, category_id, trade_mode, allow_walkthrough, allow_pets, allow_pets_eat, hide_walls, wall_thickness, floor_thickness, chat_mode, chat_weight, chat_speed, chat_distance, chat_protection, moderation_mute, moderation_kick, moderation_ban, staff_picked, public_room, created_at, updated_at, deleted_at, version`

	// createRoomSQL inserts a room record.
	createRoomSQL = `
insert into rooms (owner_player_id, owner_name, name, description, model_name, door_mode, max_users, category_id, trade_mode)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
returning ` + roomColumns

	// findRoomByIDSQL reads one active room by id.
	findRoomByIDSQL = `select ` + roomColumns + ` from rooms where id = $1 and deleted_at is null`

	// listRoomsByOwnerSQL reads active rooms by owner.
	listRoomsByOwnerSQL = `select ` + roomColumns + ` from rooms where owner_player_id = $1 and deleted_at is null order by id desc`

	// listPopularRoomsSQL reads active rooms by score and recency.
	listPopularRoomsSQL = `select ` + roomColumns + ` from rooms where deleted_at is null order by score desc, updated_at desc limit $1`

	// listHighestScoreRoomsSQL reads active rooms by score.
	listHighestScoreRoomsSQL = `select ` + roomColumns + ` from rooms where deleted_at is null order by score desc, id asc limit $1`

	// searchRoomsSQL reads active rooms matching public navigator text.
	searchRoomsSQL = `select ` + roomColumns + ` from rooms where deleted_at is null and (name ilike $1 or owner_name ilike $1 or description ilike $1 or exists (select 1 from room_tags where room_tags.room_id = rooms.id and room_tags.tag ilike $1)) order by score desc, updated_at desc limit $2`

	// softDeleteRoomSQL soft deletes one active room.
	softDeleteRoomSQL = `update rooms set deleted_at = now(), updated_at = now(), version = version + 1 where id = $1 and deleted_at is null`
)

// CreateRoomParams contains room creation data.
type CreateRoomParams struct {
	// OwnerPlayerID identifies the player that owns the room.
	OwnerPlayerID int64

	// OwnerName stores an owner name snapshot for navigator listings.
	OwnerName string

	// Name is the visible room name.
	Name string

	// Description is the visible room description.
	Description string

	// ModelName is the room layout model name.
	ModelName string

	// DoorMode describes how the room accepts entry.
	DoorMode roommodel.DoorMode

	// MaxUsers stores the maximum user count.
	MaxUsers int

	// CategoryID optionally identifies the navigator category.
	CategoryID *int64

	// TradeMode describes trading behavior.
	TradeMode roommodel.TradeMode
}

// CreateRoom creates a room record.
func (repository *Repository) CreateRoom(ctx context.Context, params CreateRoomParams) (roommodel.Room, error) {
	room, err := scanRoom(repository.executor.QueryRow(ctx, createRoomSQL, params.OwnerPlayerID, params.OwnerName, params.Name, params.Description, params.ModelName, params.DoorMode, params.MaxUsers, params.CategoryID, params.TradeMode))
	if err != nil {
		return roommodel.Room{}, fmt.Errorf("create room: %w", err)
	}

	return room, nil
}

// FindRoomByID finds an active room by id.
func (repository *Repository) FindRoomByID(ctx context.Context, id int64) (roommodel.Room, bool, error) {
	return repository.findRoom(ctx, findRoomByIDSQL, id)
}

// ListRoomsByOwner lists active rooms owned by a player.
func (repository *Repository) ListRoomsByOwner(ctx context.Context, ownerPlayerID int64) ([]roommodel.Room, error) {
	return repository.listRooms(ctx, listRoomsByOwnerSQL, ownerPlayerID)
}

// ListPopularRooms lists active rooms by occupancy-facing popularity fields.
func (repository *Repository) ListPopularRooms(ctx context.Context, limit int) ([]roommodel.Room, error) {
	return repository.listRooms(ctx, listPopularRoomsSQL, limit)
}

// ListHighestScoreRooms lists active rooms by score.
func (repository *Repository) ListHighestScoreRooms(ctx context.Context, limit int) ([]roommodel.Room, error) {
	return repository.listRooms(ctx, listHighestScoreRoomsSQL, limit)
}

// SearchRooms searches active rooms by public navigator text.
func (repository *Repository) SearchRooms(ctx context.Context, query string, limit int) ([]roommodel.Room, error) {
	return repository.listRooms2(ctx, searchRoomsSQL, "%"+query+"%", limit)
}

// SoftDeleteRoom soft deletes a room record.
func (repository *Repository) SoftDeleteRoom(ctx context.Context, id int64) (bool, error) {
	tag, err := repository.executor.Exec(ctx, softDeleteRoomSQL, id)
	if err != nil {
		return false, fmt.Errorf("soft delete room: %w", err)
	}

	return tag.RowsAffected() > 0, nil
}

// findRoom finds one room with a query.
func (repository *Repository) findRoom(ctx context.Context, query string, argument any) (roommodel.Room, bool, error) {
	room, err := scanRoom(repository.executor.QueryRow(ctx, query, argument))
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
	rows, err := repository.executor.Query(ctx, query, argument)
	if err != nil {
		return nil, fmt.Errorf("list rooms: %w", err)
	}
	defer rows.Close()

	return scanRooms(rows)
}

// listRooms2 lists rooms with two query arguments.
func (repository *Repository) listRooms2(ctx context.Context, query string, first any, second any) ([]roommodel.Room, error) {
	rows, err := repository.executor.Query(ctx, query, first, second)
	if err != nil {
		return nil, fmt.Errorf("list rooms: %w", err)
	}
	defer rows.Close()

	return scanRooms(rows)
}
