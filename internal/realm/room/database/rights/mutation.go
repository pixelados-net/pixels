package repository

import (
	"context"
	"fmt"

	rightsmodel "github.com/niflaot/pixels/internal/realm/room/control/rights/model"
)

// Grant creates rights when absent.
func (repository *Repository) Grant(ctx context.Context, roomID int64, playerID int64, actorID int64) (bool, error) {
	tag, err := repository.executorFor(ctx).Exec(ctx, `
		insert into room_rights (room_id, player_id, granted_by_player_id)
		values ($1, $2, $3)
		on conflict (room_id, player_id) do nothing`, roomID, playerID, actorID)
	if err != nil {
		return false, fmt.Errorf("grant room %d rights to player %d: %w", roomID, playerID, err)
	}

	return tag.RowsAffected() > 0, nil
}

// Revoke removes one rights holder.
func (repository *Repository) Revoke(ctx context.Context, roomID int64, playerID int64) (bool, error) {
	tag, err := repository.executorFor(ctx).Exec(ctx, `delete from room_rights where room_id=$1 and player_id=$2`, roomID, playerID)
	if err != nil {
		return false, fmt.Errorf("revoke room %d rights from player %d: %w", roomID, playerID, err)
	}

	return tag.RowsAffected() > 0, nil
}

// RevokeAll removes and returns every rights holder.
func (repository *Repository) RevokeAll(ctx context.Context, roomID int64) ([]rightsmodel.Right, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `
		delete from room_rights where room_id=$1
		returning room_id, player_id, '', granted_by_player_id, created_at`, roomID)
	if err != nil {
		return nil, fmt.Errorf("revoke all room %d rights: %w", roomID, err)
	}
	defer rows.Close()

	return scanRights(rows)
}
