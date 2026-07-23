// Package moved contains the completed room-unit movement event.
package moved

import (
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/pkg/bus"
)

// Name identifies a completed unit step after the authoritative world mutation.
const Name bus.Name = "room.unit.moved"

// Payload contains stable movement context for room capabilities.
type Payload struct {
	// RoomID identifies the containing room.
	RoomID int64
	// EntityKey identifies the moved runtime unit.
	EntityKey int64
	// PlayerID identifies a durable player and is zero for bots and pets.
	PlayerID int64
	// Kind classifies the moved room unit.
	Kind worldunit.Kind
	// Previous stores the position before the accepted step.
	Previous worldpath.Position
	// Current stores the position after the accepted step.
	Current worldpath.Position
}
