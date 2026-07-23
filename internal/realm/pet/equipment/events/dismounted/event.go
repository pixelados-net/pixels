// Package dismounted contains the pet dismounted event.
package dismounted

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the pet dismounted event.
const Name bus.Name = "pet.dismounted"

// Payload describes one room-local dismount.
type Payload struct {
	// PetID identifies the dismounted pet.
	PetID int64
	// RoomID identifies the room.
	RoomID int64
	// PlayerID identifies the former rider.
	PlayerID int64
}
