package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	rightsmodel "github.com/niflaot/pixels/internal/realm/room/control/rights/model"
)

// List lists current rights holders.
func (repository *Repository) List(ctx context.Context, roomID int64) ([]rightsmodel.Right, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `
		select rr.room_id, rr.player_id, p.username, rr.granted_by_player_id, rr.created_at
		from room_rights rr join players p on p.id=rr.player_id
		where rr.room_id=$1 order by lower(p.username), rr.player_id`, roomID)
	if err != nil {
		return nil, fmt.Errorf("list room %d rights: %w", roomID, err)
	}
	defer rows.Close()

	return scanRights(rows)
}

// Exists reports whether a player holds rights.
func (repository *Repository) Exists(ctx context.Context, roomID int64, playerID int64) (bool, error) {
	var exists bool
	err := repository.executorFor(ctx).QueryRow(ctx, `select exists(select 1 from room_rights where room_id=$1 and player_id=$2)`, roomID, playerID).Scan(&exists)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("check room %d rights for player %d: %w", roomID, playerID, err)
	}

	return exists, nil
}

// scanRights scans room rights rows.
func scanRights(rows pgx.Rows) ([]rightsmodel.Right, error) {
	rights := make([]rightsmodel.Right, 0)
	for rows.Next() {
		var right rightsmodel.Right
		if err := rows.Scan(&right.RoomID, &right.PlayerID, &right.Username, &right.GrantedByPlayerID, &right.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan room rights: %w", err)
		}
		rights = append(rights, right)
	}

	return rights, rows.Err()
}
