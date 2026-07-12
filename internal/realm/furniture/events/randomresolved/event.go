// Package randomresolved contains the furniture random-result event.
package randomresolved

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies a resolved random furniture interaction.
const Name bus.Name = "furniture.random_resolved"

// Payload describes one settled random furniture result.
type Payload struct {
	// PlayerID identifies the player that triggered the interaction.
	PlayerID int64
	// ItemID identifies the resolved furniture item.
	ItemID int64
	// RoomID identifies the containing room.
	RoomID int64
	// Result stores the selected visual state.
	Result int
}
