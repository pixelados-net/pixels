// Package model contains persistent navigator records.
package record

import "time"

// Favorite stores a player's favorited room.
type Favorite struct {
	// PlayerID identifies the player.
	PlayerID int64

	// RoomID identifies the room.
	RoomID int64

	// CreatedAt is the time the favorite was created.
	CreatedAt time.Time
}
