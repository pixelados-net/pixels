package repository

import (
	"context"
	"fmt"
	"time"
)

// Mute creates or replaces a room mute.
func (repository *Repository) Mute(ctx context.Context, roomID int64, playerID int64, endsAt time.Time) error {
	return repository.upsert(ctx, "room_mutes", roomID, playerID, endsAt)
}

// Unmute ends an active room mute.
func (repository *Repository) Unmute(ctx context.Context, roomID int64, playerID int64, now time.Time) (bool, error) {
	return repository.end(ctx, "room_mutes", roomID, playerID, now)
}

// Ban creates or replaces a room ban.
func (repository *Repository) Ban(ctx context.Context, roomID int64, playerID int64, endsAt time.Time) error {
	return repository.upsert(ctx, "room_bans", roomID, playerID, endsAt)
}

// Unban ends an active room ban.
func (repository *Repository) Unban(ctx context.Context, roomID int64, playerID int64, now time.Time) (bool, error) {
	return repository.end(ctx, "room_bans", roomID, playerID, now)
}

// upsert creates or replaces one current sanction.
func (repository *Repository) upsert(ctx context.Context, table string, roomID int64, playerID int64, endsAt time.Time) error {
	query := `insert into ` + table + ` (room_id, player_id, ends_at) values ($1,$2,$3)
		on conflict (room_id, player_id) do update set ends_at=excluded.ends_at, created_at=now(), updated_at=now()`
	if _, err := repository.executorFor(ctx).Exec(ctx, query, roomID, playerID, endsAt); err != nil {
		return fmt.Errorf("upsert %s room %d player %d: %w", table, roomID, playerID, err)
	}

	return nil
}

// end expires one active sanction.
func (repository *Repository) end(ctx context.Context, table string, roomID int64, playerID int64, now time.Time) (bool, error) {
	query := `update ` + table + ` set ends_at=$3, updated_at=$3 where room_id=$1 and player_id=$2 and ends_at>$3`
	tag, err := repository.executorFor(ctx).Exec(ctx, query, roomID, playerID, now)
	if err != nil {
		return false, fmt.Errorf("end %s room %d player %d: %w", table, roomID, playerID, err)
	}

	return tag.RowsAffected() > 0, nil
}
