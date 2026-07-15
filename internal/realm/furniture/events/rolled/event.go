// Package rolled contains the consolidated furniture roller event.
package rolled

import (
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	"github.com/niflaot/pixels/pkg/bus"
)

// Name identifies a completed roller step.
const Name bus.Name = "furniture.rolled"

// Payload describes all entities moved by one roller step.
type Payload struct {
	// RoomID identifies the active room.
	RoomID int64
	// RollerItemID identifies the roller furniture item.
	RollerItemID int64
	// ItemIDs identifies moved furniture.
	ItemIDs []int64
	// PlayerIDs identifies moved player units.
	PlayerIDs []int64
	// From stores the roller tile.
	From grid.Point
	// To stores the destination tile.
	To grid.Point
}
