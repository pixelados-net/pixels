// Package fed defines committed pet feeding events.
package fed

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies one committed pet food, drink, toy, or hand-item use.
const Name bus.Name = "pet.fed"

// Payload describes one pet care action.
type Payload struct {
	// PlayerID identifies the player who fed the pet.
	PlayerID int64
	// PetID identifies the cared-for pet.
	PetID int64
}
