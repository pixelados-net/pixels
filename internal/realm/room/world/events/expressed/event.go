// Package expressed defines room unit expression events.
package expressed

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies a room unit expression.
const Name bus.Name = "room.unit.expressed"

// Payload describes an expression.
type Payload struct {
	// RoomID identifies the active room.
	RoomID int64
	// RoomIndex identifies the room-local unit.
	RoomIndex int64
	// ExpressionID stores the emitted expression.
	ExpressionID int32
}
