package repository

import (
	"context"
	"fmt"

	roomaudit "github.com/niflaot/pixels/internal/realm/room/control/audit"
	auditmodel "github.com/niflaot/pixels/internal/realm/room/control/audit/model"
)

// RightsHistory lists matching rights history.
func (repository *Repository) RightsHistory(ctx context.Context, query roomaudit.StoreQuery) ([]auditmodel.RightsAudit, error) {
	rows, err := repository.executorFor(ctx).Query(ctx, `
		select id, room_id, player_id, actor_kind, actor_id, action, created_at
		from room_rights_audit
		where ($1::bigint is null or room_id=$1)
		  and ($2::bigint is null or player_id=$2)
		  and ($3::bigint is null or actor_id=$3)
		  and ($4::bigint is null or id<$4)
		order by id desc limit $5`, query.RoomID, query.TargetPlayerID, query.ActorPlayerID, query.Before, query.Limit)
	if err != nil {
		return nil, fmt.Errorf("query room rights audit: %w", err)
	}
	defer rows.Close()

	items := make([]auditmodel.RightsAudit, 0)
	for rows.Next() {
		var item auditmodel.RightsAudit
		if err := rows.Scan(&item.ID, &item.RoomID, &item.PlayerID, &item.ActorKind, &item.ActorID, &item.Action, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan room rights audit: %w", err)
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

// ModerationHistory lists matching moderation history.
func (repository *Repository) ModerationHistory(ctx context.Context, query roomaudit.StoreQuery) ([]auditmodel.ModerationAction, error) {
	actions := make([]string, len(query.ActionTypes))
	for index, action := range query.ActionTypes {
		actions[index] = string(action)
	}
	rows, err := repository.executorFor(ctx).Query(ctx, `
		select id, room_id, target_player_id, actor_kind, actor_id, action_type, duration_seconds, expires_at, created_at
		from room_moderation_actions
		where ($1::bigint is null or room_id=$1)
		  and ($2::bigint is null or target_player_id=$2)
		  and ($3::bigint is null or actor_id=$3)
		  and (cardinality($4::text[])=0 or action_type=any($4))
		  and ($5::bigint is null or id<$5)
		order by id desc limit $6`, query.RoomID, query.TargetPlayerID, query.ActorPlayerID, actions, query.Before, query.Limit)
	if err != nil {
		return nil, fmt.Errorf("query room moderation audit: %w", err)
	}
	defer rows.Close()

	items := make([]auditmodel.ModerationAction, 0)
	for rows.Next() {
		var item auditmodel.ModerationAction
		if err := rows.Scan(&item.ID, &item.RoomID, &item.TargetPlayerID, &item.ActorKind, &item.ActorID, &item.Action, &item.DurationSeconds, &item.ExpiresAt, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan room moderation audit: %w", err)
		}
		items = append(items, item)
	}

	return items, rows.Err()
}
