// Package deleted contains the room deleted event.
package deleted

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the room deleted event.
const Name bus.Name = "room.deleted"

// Payload describes a room deletion event.
type Payload struct {
	// RoomID identifies the room.
	RoomID int64
}
