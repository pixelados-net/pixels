// Package updated contains the room updated event.
package updated

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the room updated event.
const Name bus.Name = "room.updated"

// Payload describes a room update event.
type Payload struct {
	// RoomID identifies the room.
	RoomID int64
}
