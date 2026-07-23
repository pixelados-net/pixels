// Package created contains the room created event.
package created

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the room created event.
const Name bus.Name = "room.created"

// Payload describes a room creation event.
type Payload struct {
	// RoomID identifies the room.
	RoomID int64

	// OwnerPlayerID identifies the room owner.
	OwnerPlayerID int64
}
