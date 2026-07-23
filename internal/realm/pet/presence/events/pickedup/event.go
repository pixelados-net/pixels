// Package pickedup contains the pet picked-up event.
package pickedup

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the pet picked-up event.
const Name bus.Name = "pet.picked_up"

// Payload describes one committed room pickup.
type Payload struct {
	// PetID identifies the picked-up pet.
	PetID int64
	// RoomID identifies the source room.
	RoomID int64
	// ActorPlayerID identifies the actor.
	ActorPlayerID int64
}
