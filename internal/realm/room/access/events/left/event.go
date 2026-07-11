// Package left contains the room left event.
package left

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the room left event.
const Name bus.Name = "room.left"

// Payload describes a player room leave event.
type Payload struct {
	// PlayerID identifies the player.
	PlayerID int64

	// RoomID identifies the room.
	RoomID int64
}
