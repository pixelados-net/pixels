// Package respected contains the pet respected event.
package respected

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the pet respected event.
const Name bus.Name = "pet.respected"

// Payload describes one committed daily respect.
type Payload struct {
	// PetID identifies the respected pet.
	PetID int64
	// OwnerPlayerID identifies the pet owner receiving progression.
	OwnerPlayerID int64
	// ActorPlayerID identifies the respecting player.
	ActorPlayerID int64
	// Respect stores the new accumulated value.
	Respect int32
}
