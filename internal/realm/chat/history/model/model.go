// Package model contains durable chat history records.
package model

import "time"

// Entry describes one durable delivered chat message.
type Entry struct {
	// ID identifies the persisted message.
	ID int64 `json:"id"`
	// RoomID identifies the room.
	RoomID int64 `json:"roomId"`
	// PlayerID identifies the speaker.
	PlayerID int64 `json:"playerId"`
	// TargetPlayerID optionally identifies a whisper recipient.
	TargetPlayerID *int64 `json:"targetPlayerId,omitempty"`
	// Kind stores talk, shout, or whisper.
	Kind string `json:"kind"`
	// Message stores text visible to recipients.
	Message string `json:"message"`
	// Censored reports whether filtering changed the text.
	Censored bool `json:"censored"`
	// CreatedAt stores delivery time.
	CreatedAt time.Time `json:"createdAt"`
}

// Query contains keyset history filters.
type Query struct {
	// RoomID optionally restricts results to one room.
	RoomID *int64
	// PlayerID optionally restricts results to one speaker.
	PlayerID *int64
	// Before returns ids lower than this cursor.
	Before *int64
	// Limit stores the requested page size.
	Limit int
}

// Normalize bounds history query page size.
func (query Query) Normalize() Query {
	if query.Limit <= 0 {
		query.Limit = 50
	}
	if query.Limit > 200 {
		query.Limit = 200
	}

	return query
}
