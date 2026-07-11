// Package muteallchanged contains the room mute-all changed event.
package muteallchanged

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the room mute-all changed event.
const Name bus.Name = "room.mute_all_changed"

// Payload describes an active room mute-all transition.
type Payload struct {
	// RoomID identifies the room.
	RoomID int64
	// ActorID identifies the actor.
	ActorID int64
	// Muted stores the resulting state.
	Muted bool
}
