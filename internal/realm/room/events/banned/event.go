// Package banned defines the room moderation banned event.
package banned

import (
	"time"

	"github.com/niflaot/pixels/pkg/bus"
)

const (
	// Name identifies a committed room ban.
	Name bus.Name = "room.moderation_banned"
)

// Payload describes a committed room ban.
type Payload struct {
	// RoomID identifies the room.
	RoomID int64
	// TargetPlayerID identifies the banned player.
	TargetPlayerID int64
	// ActorID identifies the moderator.
	ActorID int64
	// DurationSeconds stores the ban duration.
	DurationSeconds int64
	// ExpiresAt stores when the ban expires.
	ExpiresAt time.Time
}
