// Package walkedoff contains the furniture item walked-off event, planned for future interactions.
package walkedoff

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the furniture item walked-off event.
const Name bus.Name = "furniture.walked_off"

// Payload describes a unit walking off a furniture item's footprint.
type Payload struct {
	// PlayerID identifies the unit's owning player.
	PlayerID int64

	// ItemID identifies the furniture item walked off.
	ItemID int64

	// RoomID identifies the room the item is placed in.
	RoomID int64
}
