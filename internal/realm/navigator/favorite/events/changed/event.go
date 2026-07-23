// Package favoritechanged contains the navigator favorite changed event.
package favoritechanged

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the navigator favorite changed event.
const Name bus.Name = "navigator.favorite_changed"

// Payload describes a navigator favorite change.
type Payload struct {
	// PlayerID identifies the player.
	PlayerID int64

	// RoomID identifies the room.
	RoomID int64

	// Added reports whether the favorite was added.
	Added bool
}
