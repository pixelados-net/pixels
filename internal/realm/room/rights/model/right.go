// Package model contains persistent room rights records.
package model

import "time"

// Right describes one player's current room build rights.
type Right struct {
	// RoomID identifies the room.
	RoomID int64
	// PlayerID identifies the rights holder.
	PlayerID int64
	// Username stores the current player username for protocol lists.
	Username string
	// GrantedByPlayerID identifies the player who granted the rights.
	GrantedByPlayerID int64
	// CreatedAt stores when rights were granted.
	CreatedAt time.Time
}
