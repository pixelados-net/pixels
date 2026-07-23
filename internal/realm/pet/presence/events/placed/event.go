// Package placed contains the pet placed event.
package placed

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the pet placed event.
const Name bus.Name = "pet.placed"

// Payload describes one committed room placement.
type Payload struct {
	// PetID identifies the placed pet.
	PetID int64
	// RoomID identifies the destination room.
	RoomID int64
	// ActorPlayerID identifies the actor.
	ActorPlayerID int64
}
