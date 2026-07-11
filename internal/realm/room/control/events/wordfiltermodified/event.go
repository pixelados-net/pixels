// Package wordfiltermodified contains the room word filter modified event.
package wordfiltermodified

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the room word filter modified event.
const Name bus.Name = "room.word_filter_modified"

// Payload describes one room word filter mutation.
type Payload struct {
	// RoomID identifies the room.
	RoomID int64
	// ActorID identifies the actor.
	ActorID int64
	// Added reports whether the word was added.
	Added bool
	// Word stores the normalized filter word.
	Word string
}
