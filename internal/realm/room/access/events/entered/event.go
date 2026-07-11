// Package entered contains the room entered event.
package entered

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the room entered event.
const Name bus.Name = "room.entered"

// Payload describes a player room entry event.
type Payload struct {
	// PlayerID identifies the player.
	PlayerID int64

	// RoomID identifies the room.
	RoomID int64
}
