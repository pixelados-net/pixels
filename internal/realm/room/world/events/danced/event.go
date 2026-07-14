// Package danced defines room unit dance events.
package danced

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies a room unit dance change.
const Name bus.Name = "room.unit.danced"

// Payload describes a dance change.
type Payload struct {
	// RoomID identifies the active room.
	RoomID int64
	// RoomIndex identifies the room-local unit.
	RoomIndex int64
	// DanceID stores the selected dance.
	DanceID int32
}
