// Package harvested contains the pet monsterplant-harvested event.
package harvested

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the monsterplant harvested event.
const Name bus.Name = "pet.harvested"

// Payload describes one committed plant consumption.
type Payload struct {
	// PetID identifies the consumed plant.
	PetID int64
	// OwnerPlayerID identifies the reward recipient.
	OwnerPlayerID int64
	// State identifies harvest or compost completion.
	State string
}
