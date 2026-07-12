// Package walkedon contains the furniture item walked-on event.
package walkedon

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the furniture item walked-on event.
const Name bus.Name = "furniture.walked_on"

// Payload describes a unit walking onto a furniture item's footprint.
type Payload struct {
	// PlayerID identifies the unit's owning player.
	PlayerID int64

	// ItemID identifies the furniture item walked onto.
	ItemID int64

	// RoomID identifies the room the item is placed in.
	RoomID int64
}
