// Package player contains immutable player values shared across plugin SDK capabilities.
package player

// Player is an immutable plugin-facing player snapshot.
type Player struct {
	// ID identifies the durable player.
	ID int64
	// Username stores the current player name.
	Username string
	// RoomID identifies the current room or zero when outside a room.
	RoomID int64
	// Online reports whether the player has an authenticated live session.
	Online bool
}
