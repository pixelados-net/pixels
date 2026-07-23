// Package created contains the pet created event.
package created

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the pet created event.
const Name bus.Name = "pet.created"

// Payload describes one committed pet creation.
type Payload struct {
	// PetID identifies the created pet.
	PetID int64
	// OwnerPlayerID identifies the receiving owner.
	OwnerPlayerID int64
	// TypeID identifies the created species.
	TypeID int32
}
