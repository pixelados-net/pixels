// Package leveled contains the pet leveled event.
package leveled

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the pet leveled event.
const Name bus.Name = "pet.leveled"

// Payload describes one committed level transition.
type Payload struct {
	// PetID identifies the leveled pet.
	PetID int64
	// OwnerPlayerID identifies the player receiving progression.
	OwnerPlayerID int64
	// PreviousLevel stores the prior level.
	PreviousLevel int32
	// Level stores the resulting level.
	Level int32
}
