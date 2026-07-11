// Package unmuted defines the room moderation unmuted event.
package unmuted

import "github.com/niflaot/pixels/pkg/bus"

const (
	// Name identifies a committed room unmute.
	Name bus.Name = "room.moderation_unmuted"
)

// Payload describes a committed room unmute.
type Payload struct {
	// RoomID identifies the room.
	RoomID int64
	// TargetPlayerID identifies the unmuted player.
	TargetPlayerID int64
	// ActorID identifies the moderator.
	ActorID int64
}
