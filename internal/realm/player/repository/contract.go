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

// AdminWriter mutates complete player records for administrative workflows.
type AdminWriter interface {
	// UpdatePlayer updates one active player identity with optimistic locking.
	UpdatePlayer(ctx context.Context, params UpdatePlayerParams) (playermodel.Player, bool, error)

	// UpdateProfile updates one complete player profile with optimistic locking.
	UpdateProfile(ctx context.Context, params UpdateProfileParams) (playermodel.Profile, bool, error)

	// SoftDeletePlayer marks one active player deleted with optimistic locking.
	SoftDeletePlayer(ctx context.Context, playerID int64, expectedVersion int64) (bool, error)
}

// ClubWriter writes derived player club entitlement state.
type ClubWriter interface {
	// UpdateClub updates the derived player club entitlement.
	UpdateClub(ctx context.Context, playerID int64, club playermodel.Club) error
}

// TradeWriter updates durable direct-trade eligibility.
type TradeWriter interface {
	// UpdateAllowTrade updates a player's trade eligibility.
	UpdateAllowTrade(ctx context.Context, playerID int64, allow bool) (bool, error)
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

	// UpdatePrivacy persists messenger privacy fields.
	UpdatePrivacy(ctx context.Context, playerID int64, params PrivacyParams) (playermodel.Profile, error)
}

// Store reads and writes player persistence records.
type Store interface {
	PlayerReader
	PlayerWriter
	ProfileReader
	ProfileWriter

	// WithinTransaction runs player creation work atomically.
	WithinTransaction(ctx context.Context, work func(context.Context) error) error
}

// storeAssertion verifies Repository implements Store.
var storeAssertion Store = (*Repository)(nil)
