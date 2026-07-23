package model

import (
	"time"

	moderationmodel "github.com/niflaot/pixels/internal/realm/room/control/moderation/model"
)

// ModerationAction stores one append-only room moderation action.
type ModerationAction struct {
	// ID identifies the audit row.
	ID int64 `json:"id"`
	// RoomID identifies the room.
	RoomID int64 `json:"roomId"`
	// TargetPlayerID identifies the affected player.
	TargetPlayerID int64 `json:"targetPlayerId"`
	// ActorKind identifies the source family.
	ActorKind string `json:"actorKind"`
	// ActorID optionally identifies the source player.
	ActorID *int64 `json:"actorId,omitempty"`
	// Action identifies the moderation operation.
	Action moderationmodel.Action `json:"action"`
	// DurationSeconds optionally stores the sanction duration.
	DurationSeconds *int64 `json:"durationSeconds,omitempty"`
	// ExpiresAt optionally stores sanction expiry.
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
	// CreatedAt stores when the action occurred.
	CreatedAt time.Time `json:"createdAt"`
}
