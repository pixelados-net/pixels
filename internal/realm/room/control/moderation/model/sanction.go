package model

import "time"

// Sanction describes one current room mute or ban.
type Sanction struct {
	// RoomID identifies the room.
	RoomID int64 `json:"roomId"`
	// PlayerID identifies the affected player.
	PlayerID int64 `json:"playerId"`
	// Username stores the current player username.
	Username string `json:"username"`
	// EndsAt stores when the sanction expires.
	EndsAt time.Time `json:"endsAt"`
	// CreatedAt stores when the current sanction began.
	CreatedAt time.Time `json:"createdAt"`
	// UpdatedAt stores when the sanction was last changed.
	UpdatedAt time.Time `json:"updatedAt"`
}
