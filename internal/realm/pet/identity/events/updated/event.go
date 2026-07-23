// Package updated contains the pet updated event.
package updated

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the pet updated event.
const Name bus.Name = "pet.updated"

// Payload describes one committed pet mutation.
type Payload struct {
	// PetID identifies the updated pet.
	PetID int64
	// Version stores the new optimistic version.
	Version int64
}
