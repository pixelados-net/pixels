// Package votecast defines the room vote cast event.
package votecast

import "github.com/niflaot/pixels/pkg/bus"

const (
	// Name identifies a committed room vote.
	Name bus.Name = "room.vote_cast"
)

// Payload describes a committed room vote.
type Payload struct {
	// RoomID identifies the rated room.
	RoomID int64
	// PlayerID identifies the voting player.
	PlayerID int64
	// Score stores the resulting room score.
	Score int
}
