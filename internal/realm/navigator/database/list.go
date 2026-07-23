package database

import (
	"context"
	"fmt"

	navmodel "github.com/niflaot/pixels/internal/realm/navigator/record"
)

const (
	// addFavoriteSQL inserts a player favorite.
	addFavoriteSQL = `insert into room_favorites(player_id,room_id) select $1,r.id from rooms r where r.id=$2 and r.deleted_at is null and not r.is_bundle_template and ($4 or (select count(*) from room_favorites where player_id=$1)<$3) on conflict do nothing`
	// removeFavoriteSQL deletes a player favorite.
	removeFavoriteSQL = `delete from room_favorites where player_id = $1 and room_id = $2`
	// listFavoriteRoomIDsSQL reads active favorite room identifiers through one bounded join.
	listFavoriteRoomIDsSQL = `select favorites.room_id from room_favorites favorites join rooms on rooms.id=favorites.room_id where favorites.player_id=$1 and rooms.deleted_at is null and not rooms.is_bundle_template order by favorites.created_at desc`
	// liftedRoomColumns contains the shared lifted room select list.
	liftedRoomColumns = `id, room_id, area_id, image, caption, order_num, starts_at, ends_at, created_at, updated_at, deleted_at, version`
	// listLiftedRoomsSQL reads active lifted rooms.
	listLiftedRoomsSQL = `select ` + liftedRoomColumns + ` from navigator_lifted_rooms where deleted_at is null and (starts_at is null or starts_at <= now()) and (ends_at is null or ends_at > now()) order by order_num asc, id asc`
)

// AddFavorite adds a favorite room for a player.
func (repository *Repository) AddFavorite(ctx context.Context, playerID int64, roomID int64, limit int32, unlimited bool) error {
	result, err := repository.executor.Exec(ctx, addFavoriteSQL, playerID, roomID, limit, unlimited)
	if err != nil {
		return fmt.Errorf("add room favorite: %w", err)
	}
	if result.RowsAffected() == 0 {
		return navmodel.ErrFavoriteUnavailable
	}
	return nil
}

// RemoveFavorite removes a favorite room for a player.
func (repository *Repository) RemoveFavorite(ctx context.Context, playerID int64, roomID int64) error {
	if _, err := repository.executor.Exec(ctx, removeFavoriteSQL, playerID, roomID); err != nil {
		return fmt.Errorf("remove room favorite: %w", err)
	}
	return nil
}

// ListFavoriteRoomIDs lists favorite room identifiers for a player.
func (repository *Repository) ListFavoriteRoomIDs(ctx context.Context, playerID int64) ([]int64, error) {
	rows, err := repository.executor.Query(ctx, listFavoriteRoomIDsSQL, playerID)
	if err != nil {
		return nil, fmt.Errorf("list room favorites: %w", err)
	}
	defer rows.Close()
	return scanFavoriteRoomIDs(rows)
}

// ListLiftedRooms lists currently active lifted rooms.
func (repository *Repository) ListLiftedRooms(ctx context.Context) ([]navmodel.LiftedRoom, error) {
	rows, err := repository.executor.Query(ctx, listLiftedRoomsSQL)
	if err != nil {
		return nil, fmt.Errorf("list lifted rooms: %w", err)
	}
	defer rows.Close()
	return scanLiftedRooms(rows)
}
