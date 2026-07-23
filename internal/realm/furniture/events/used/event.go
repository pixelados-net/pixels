// Package used contains the furniture item used event.
package used

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the furniture item used event.
const Name bus.Name = "furniture.used"

// Payload describes a unit toggling or using a furniture item.
type Payload struct {
	// PlayerID identifies the actor using the item.
	PlayerID int64

	// ItemID identifies the furniture item used.
	ItemID int64

	// RoomID identifies the room the item is placed in.
	RoomID int64
}
