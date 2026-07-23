// Package surfaceapplied contains committed room surface decoration events.
package surfaceapplied

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies a committed room surface decoration.
const Name bus.Name = "furniture.surface_applied"

// Payload describes one floor, wallpaper, or landscape application.
type Payload struct {
	// PlayerID identifies the actor.
	PlayerID int64
	// RoomID identifies the decorated room.
	RoomID int64
	// Surface stores floor, wallpaper, or landscape.
	Surface string
}
