package votes

import (
	"context"

	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
)

// Store persists and reads room votes.
type Store interface {
	// Cast atomically inserts a vote and increments the room score once.
	Cast(context.Context, int64, int64) (Mutation, error)
	// HasVote reports whether one player voted for one room.
	HasVote(context.Context, int64, int64) (bool, error)
	// Existing returns voters present in a supplied player id set.
	Existing(context.Context, int64, []int64) (map[int64]struct{}, error)
	// List returns durable votes matching a query.
	List(context.Context, Query) ([]Vote, error)
}

// RoomFinder reads durable room metadata.
type RoomFinder interface {
	// FindByID finds a room by id.
	FindByID(context.Context, int64) (roommodel.Room, bool, error)
}

// Reader reads vote state.
type Reader interface {
	// State reads one player's room score and eligibility.
	State(context.Context, int64, int64) (State, error)
	// List returns durable room votes.
	List(context.Context, Query) ([]Vote, error)
}

// Manager reads and mutates room votes.
type Manager interface {
	Reader
	// Cast permanently upvotes a room once per player.
	Cast(context.Context, int64, int64) (Mutation, error)
}
