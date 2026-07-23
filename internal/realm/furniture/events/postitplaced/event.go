// Package postitplaced contains committed post-it placement events.
package postitplaced

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies one committed post-it placement.
const Name bus.Name = "furniture.postit_placed"

// Payload describes the actor and receiving room owner.
type Payload struct {
	// PlayerID identifies the actor who placed the note.
	PlayerID int64
	// RoomOwnerID identifies the owner receiving the note.
	RoomOwnerID int64
	// RoomID identifies the decorated room.
	RoomID int64
	// ItemID identifies the placed post-it.
	ItemID int64
}
