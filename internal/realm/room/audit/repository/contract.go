// Package repository persists room audit history.
package repository

import (
	"context"

	auditmodel "github.com/niflaot/pixels/internal/realm/room/audit/model"
	moderationmodel "github.com/niflaot/pixels/internal/realm/room/moderation/model"
)

// Query contains indexed audit filters.
type Query struct {
	// RoomID optionally filters one room.
	RoomID *int64
	// TargetPlayerID optionally filters an affected player.
	TargetPlayerID *int64
	// ActorPlayerID optionally filters an acting player.
	ActorPlayerID *int64
	// ActionTypes optionally filters moderation action types.
	ActionTypes []moderationmodel.Action
	// Before optionally filters ids below a cursor.
	Before *int64
	// Limit caps returned rows.
	Limit int
}

// Store persists and reads room audit records.
type Store interface {
	// InsertRights appends one rights audit row.
	InsertRights(ctx context.Context, entry auditmodel.RightsAudit) error
	// InsertModeration appends one moderation audit row.
	InsertModeration(ctx context.Context, entry auditmodel.ModerationAction) error
	// RightsHistory lists matching rights history.
	RightsHistory(ctx context.Context, query Query) ([]auditmodel.RightsAudit, error)
	// ModerationHistory lists matching moderation history.
	ModerationHistory(ctx context.Context, query Query) ([]auditmodel.ModerationAction, error)
}

// storeAssertion verifies Repository implements Store.
var storeAssertion Store = (*Repository)(nil)
