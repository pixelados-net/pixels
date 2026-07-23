// Package unbanned defines the room moderation unbanned event.
package unbanned

import "github.com/niflaot/pixels/pkg/bus"

const (
	// Name identifies a committed room unban.
	Name bus.Name = "room.moderation_unbanned"
)

// Payload describes a committed room unban.
type Payload struct {
	// RoomID identifies the room.
	RoomID int64
	// TargetPlayerID identifies the unbanned player.
	TargetPlayerID int64
	// ActorID identifies the moderator.
	ActorID int64
}
