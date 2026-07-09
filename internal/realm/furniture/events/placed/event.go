// Package placed contains the furniture item placed event.
package placed

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the furniture item placed event.
const Name bus.Name = "furniture.placed"

// Payload describes a furniture item placement event.
type Payload struct {
	// PlayerID identifies the actor who placed the item.
	PlayerID int64

	// ItemID identifies the placed furniture item.
	ItemID int64

	// DefinitionID identifies the item's furniture definition.
	DefinitionID int64

	// RoomID identifies the room the item was placed into.
	RoomID int64

	// X stores the placed floor tile x coordinate.
	X int

	// Y stores the placed floor tile y coordinate.
	Y int

	// Rotation stores the placed instance rotation.
	Rotation int
}
