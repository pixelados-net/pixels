package repository

import (
	"context"

	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
)

// PlayerReader reads player identity records.
type PlayerReader interface {
	// FindPlayerByID finds an active player by id.
	FindPlayerByID(ctx context.Context, id int64) (playermodel.Player, bool, error)

	// FindPlayerByUsername finds an active player by username.
	FindPlayerByUsername(ctx context.Context, username string) (playermodel.Player, bool, error)
}

// PlayerWriter writes player identity records.
type PlayerWriter interface {
	// CreatePlayer creates a player identity record.
	CreatePlayer(ctx context.Context, params CreatePlayerParams) (playermodel.Player, error)
}

// ProfileReader reads player profile records.
type ProfileReader interface {
	// FindProfileByPlayerID finds a profile by player id.
	FindProfileByPlayerID(ctx context.Context, playerID int64) (playermodel.Profile, bool, error)
}

// ProfileWriter writes player profile records.
type ProfileWriter interface {
	// CreateProfile creates a player profile record.
	CreateProfile(ctx context.Context, params CreateProfileParams) (playermodel.Profile, error)

	// UpdateBubbleStyle persists one validated chat bubble selection.
	UpdateBubbleStyle(ctx context.Context, playerID int64, bubbleStyle int32) (playermodel.Profile, error)
}

// Store reads and writes player persistence records.
type Store interface {
	PlayerReader
	PlayerWriter
	ProfileReader
	ProfileWriter
}

// storeAssertion verifies Repository implements Store.
var storeAssertion Store = (*Repository)(nil)
