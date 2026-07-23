// Package treated defines committed monsterplant treatment events.
package treated

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies one committed monsterplant treatment.
const Name bus.Name = "plant.treated"

// Payload describes one monsterplant treatment.
type Payload struct {
	// PlayerID identifies the plant owner.
	PlayerID int64
	// PetID identifies the monsterplant.
	PetID int64
}
