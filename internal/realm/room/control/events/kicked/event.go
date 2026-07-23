// Package kicked defines the room moderation kicked event.
package kicked

import "github.com/niflaot/pixels/pkg/bus"

const (
	// Name identifies a committed room kick.
	Name bus.Name = "room.moderation_kicked"
)

// Payload describes a committed room kick.
type Payload struct {
	// RoomID identifies the room.
	RoomID int64
	// TargetPlayerID identifies the removed player.
	TargetPlayerID int64
	// ActorID identifies the moderator.
	ActorID int64
}
