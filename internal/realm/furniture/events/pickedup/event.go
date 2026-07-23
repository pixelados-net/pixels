// Package pickedup contains the furniture item picked up event.
package pickedup

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the furniture item picked up event.
const Name bus.Name = "furniture.picked_up"

// Payload describes a furniture item pickup event.
type Payload struct {
	// PlayerID identifies the actor who picked up the item.
	PlayerID int64

	// ItemID identifies the picked up furniture item.
	ItemID int64

	// RoomID identifies the room the item was removed from.
	RoomID int64
}
