// Package moved contains the furniture item moved or rotated event.
package moved

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the furniture item moved event.
const Name bus.Name = "furniture.moved"

// Payload describes a furniture item reposition or rotation event.
type Payload struct {
	// PlayerID identifies the actor who moved the item.
	PlayerID int64

	// ItemID identifies the moved furniture item.
	ItemID int64

	// RoomID identifies the room the item is placed in.
	RoomID int64

	// X stores the destination floor tile x coordinate.
	X int

	// Y stores the destination floor tile y coordinate.
	Y int

	// Rotation stores the destination floor instance rotation.
	Rotation int
}
