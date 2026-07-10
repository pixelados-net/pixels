// Package muted defines the room moderation muted event.
package muted

import (
	"time"

	"github.com/niflaot/pixels/pkg/bus"
)

const (
	// Name identifies a committed room mute.
	Name bus.Name = "room.moderation_muted"
)

// Payload describes a committed room mute.
type Payload struct {
	// RoomID identifies the room.
	RoomID int64
	// TargetPlayerID identifies the muted player.
	TargetPlayerID int64
	// ActorID identifies the moderator.
	ActorID int64
	// DurationSeconds stores the mute duration.
	DurationSeconds int64
	// ExpiresAt stores when the mute expires.
	ExpiresAt time.Time
}
