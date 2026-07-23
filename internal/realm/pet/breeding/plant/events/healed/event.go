// Package healed defines committed monsterplant revival events.
package healed

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies one committed monsterplant revival.
const Name bus.Name = "plant.healed"

// Payload describes one monsterplant revival.
type Payload struct {
	// PlayerID identifies the plant owner.
	PlayerID int64
	// PetID identifies the monsterplant.
	PetID int64
}
