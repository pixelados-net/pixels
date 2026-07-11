// Package occupancychanged contains the room occupancy changed event.
package occupancychanged

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the room occupancy changed event.
const Name bus.Name = "room.occupancy_changed"

// Payload describes a room occupancy change.
type Payload struct {
	// RoomID identifies the room.
	RoomID int64

	// CategoryID optionally identifies the room category.
	CategoryID *int64

	// Count stores the active occupancy count.
	Count int

	// MaxUsers stores the maximum occupancy.
	MaxUsers int

	// PlayerIDs stores active player ids.
	PlayerIDs []int64
}
