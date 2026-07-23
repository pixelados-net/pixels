package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	moderationmodel "github.com/niflaot/pixels/internal/realm/room/control/moderation/model"
)

// IsMuted reports whether a room mute is active.
func (repository *Repository) IsMuted(ctx context.Context, roomID int64, playerID int64, now time.Time) (bool, error) {
	return repository.exists(ctx, "room_mutes", roomID, playerID, now)
}

// IsBanned reports whether a room ban is active.
func (repository *Repository) IsBanned(ctx context.Context, roomID int64, playerID int64, now time.Time) (bool, error) {
	return repository.exists(ctx, "room_bans", roomID, playerID, now)
}

// ListMutes lists active room mutes.
func (repository *Repository) ListMutes(ctx context.Context, roomID int64, now time.Time) ([]moderationmodel.Sanction, error) {
	return repository.list(ctx, "room_mutes", roomID, now)
}

// ListBans lists active room bans.
func (repository *Repository) ListBans(ctx context.Context, roomID int64, now time.Time) ([]moderationmodel.Sanction, error) {
	return repository.list(ctx, "room_bans", roomID, now)
}

// exists reports active sanction state.
func (repository *Repository) exists(ctx context.Context, table string, roomID int64, playerID int64, now time.Time) (bool, error) {
	var exists bool
	query := `select exists(select 1 from ` + table + ` where room_id=$1 and player_id=$2 and ends_at>$3)`
	if err := repository.executorFor(ctx).QueryRow(ctx, query, roomID, playerID, now).Scan(&exists); err != nil {
		return false, fmt.Errorf("check %s room %d player %d: %w", table, roomID, playerID, err)
	}

	return exists, nil
}

// list lists active sanctions with usernames.
func (repository *Repository) list(ctx context.Context, table string, roomID int64, now time.Time) ([]moderationmodel.Sanction, error) {
	query := `select s.room_id, s.player_id, p.username, s.ends_at, s.created_at, s.updated_at from ` + table + ` s
		join players p on p.id=s.player_id where s.room_id=$1 and s.ends_at>$2 order by s.ends_at, s.player_id`
	rows, err := repository.executorFor(ctx).Query(ctx, query, roomID, now)
	if err != nil {
		return nil, fmt.Errorf("list %s room %d: %w", table, roomID, err)
	}
	defer rows.Close()

	return scanSanctions(rows)
}

// scanSanctions scans current sanction rows.
func scanSanctions(rows pgx.Rows) ([]moderationmodel.Sanction, error) {
	items := make([]moderationmodel.Sanction, 0)
	for rows.Next() {
		var item moderationmodel.Sanction
		if err := rows.Scan(&item.RoomID, &item.PlayerID, &item.Username, &item.EndsAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan room sanction: %w", err)
		}
		items = append(items, item)
	}

	return items, rows.Err()
}
