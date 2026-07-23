// Package idlechanged defines room unit AFK events.
package idlechanged

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies a room unit idle change.
const Name bus.Name = "room.unit.idle_changed"

// Payload describes an idle change.
type Payload struct {
	// RoomID identifies the active room.
	RoomID int64
	// RoomIndex identifies the room-local unit.
	RoomIndex int64
	// Idle stores the new AFK state.
	Idle bool
}
