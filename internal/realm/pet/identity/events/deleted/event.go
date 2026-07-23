// Package deleted contains the pet deleted event.
package deleted

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the pet deleted event.
const Name bus.Name = "pet.deleted"

// Payload describes one committed pet deletion.
type Payload struct {
	// PetID identifies the deleted pet.
	PetID int64
	// OwnerPlayerID identifies its last owner.
	OwnerPlayerID int64
}
