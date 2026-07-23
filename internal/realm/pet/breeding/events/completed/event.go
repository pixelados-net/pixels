// Package completed contains the pet breeding-completed event.
package completed

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the pet breeding-completed event.
const Name bus.Name = "pet.breeding_completed"

// Payload describes one committed offspring grant.
type Payload struct {
	// NestItemID identifies the breeding nest.
	NestItemID int64
	// ParentOneID identifies the first parent.
	ParentOneID int64
	// ParentTwoID identifies the second parent.
	ParentTwoID int64
	// OffspringID identifies the granted offspring.
	OffspringID int64
}
