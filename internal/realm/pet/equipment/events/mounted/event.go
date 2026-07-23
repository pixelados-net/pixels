// Package mounted contains the pet mounted event.
package mounted

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the pet mounted event.
const Name bus.Name = "pet.mounted"

// Payload describes one room-local mount.
type Payload struct {
	// PetID identifies the mounted pet.
	PetID int64
	// RoomID identifies the room.
	RoomID int64
	// PlayerID identifies the rider.
	PlayerID int64
}
