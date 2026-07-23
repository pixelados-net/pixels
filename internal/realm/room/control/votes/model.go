// Package votes manages permanent room upvotes.
package votes

import (
	"errors"
	"time"
)

var (
	// ErrInvalidRoomID reports a malformed room id.
	ErrInvalidRoomID = errors.New("invalid vote room id")
	// ErrInvalidPlayerID reports a malformed player id.
	ErrInvalidPlayerID = errors.New("invalid vote player id")
	// ErrRoomNotFound reports a missing room.
	ErrRoomNotFound = errors.New("vote room not found")
)

const (
	// DefaultLimit stores the default administrative vote page size.
	DefaultLimit = 50
	// MaxLimit stores the maximum administrative vote page size.
	MaxLimit = 200
)

// Mutation contains one atomic vote result.
type Mutation struct {
	// Score stores the resulting room score.
	Score int
	// Inserted reports whether a new vote was persisted.
	Inserted bool
}

// State contains a player's room rating projection.
type State struct {
	// Score stores the current room score.
	Score int
	// CanVote reports whether the player may cast a vote.
	CanVote bool
	// Voted reports whether the player already voted.
	Voted bool
}

// Vote contains one durable room vote.
type Vote struct {
	// RoomID identifies the rated room.
	RoomID int64
	// PlayerID identifies the voter.
	PlayerID int64
	// CreatedAt stores when the vote was cast.
	CreatedAt time.Time
}

// Query filters administrative vote reads.
type Query struct {
	// RoomID identifies the rated room.
	RoomID int64
	// PlayerID optionally filters one voter.
	PlayerID *int64
	// Before optionally limits results to earlier votes.
	Before *time.Time
	// Limit bounds returned records.
	Limit int
}

// Normalize validates and bounds a vote query.
func (query Query) Normalize() (Query, error) {
	if query.RoomID <= 0 {
		return Query{}, ErrInvalidRoomID
	}
	if query.PlayerID != nil && *query.PlayerID <= 0 {
		return Query{}, ErrInvalidPlayerID
	}
	if query.Limit <= 0 {
		query.Limit = DefaultLimit
	}
	if query.Limit > MaxLimit {
		query.Limit = MaxLimit
	}

	return query, nil
}
