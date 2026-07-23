package repository

import (
	"context"
	"fmt"

	auditmodel "github.com/niflaot/pixels/internal/realm/room/control/audit/model"
)

// InsertRights appends one rights audit row.
func (repository *Repository) InsertRights(ctx context.Context, entry auditmodel.RightsAudit) error {
	_, err := repository.executorFor(ctx).Exec(ctx, `
		insert into room_rights_audit (room_id, player_id, actor_kind, actor_id, action, created_at)
		values ($1,$2,$3,$4,$5,$6)`, entry.RoomID, entry.PlayerID, entry.ActorKind, entry.ActorID, entry.Action, entry.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert room %d rights audit for player %d: %w", entry.RoomID, entry.PlayerID, err)
	}

	return nil
}

// InsertModeration appends one moderation audit row.
func (repository *Repository) InsertModeration(ctx context.Context, entry auditmodel.ModerationAction) error {
	_, err := repository.executorFor(ctx).Exec(ctx, `
		insert into room_moderation_actions
		(room_id, target_player_id, actor_kind, actor_id, action_type, duration_seconds, expires_at, created_at)
		values ($1,$2,$3,$4,$5,$6,$7,$8)`, entry.RoomID, entry.TargetPlayerID, entry.ActorKind, entry.ActorID, entry.Action, entry.DurationSeconds, entry.ExpiresAt, entry.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert room %d moderation audit for player %d: %w", entry.RoomID, entry.TargetPlayerID, err)
	}

	return nil
}
