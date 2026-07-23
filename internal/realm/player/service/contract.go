package service

import (
	"context"

	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
)

// Creator creates player identity and profile records.
type Creator interface {
	// Create creates a player with a profile.
	Create(ctx context.Context, params CreateParams) (Record, error)
}

// Finder reads player identity and profile records.
type Finder interface {
	// FindByID finds a player by id.
	FindByID(ctx context.Context, id int64) (Record, bool, error)

	// FindByUsername finds a player by username.
	FindByUsername(ctx context.Context, username string) (Record, bool, error)
}

// ClubWriter updates derived club entitlement fields.
type ClubWriter interface {
	// SetClub updates one player's derived club entitlement.
	SetClub(ctx context.Context, playerID int64, club playermodel.Club) error
}

// Manager creates and reads player records.
type Manager interface {
	Creator
	Finder
	// UpdatePrivacy persists messenger privacy fields.
	UpdatePrivacy(ctx context.Context, playerID int64, params PrivacyParams) (Record, error)
}

// AdminManager exposes complete protected player administration behavior.
type AdminManager interface {
	Manager
	// Update applies one partial player identity and profile mutation.
	Update(ctx context.Context, playerID int64, params UpdateParams) (Record, error)
	// SoftDelete marks one player deleted so active lookups and future logins reject it.
	SoftDelete(ctx context.Context, playerID int64) error
}

// PrivacyParams stores a complete messenger privacy replacement.
type PrivacyParams struct {
	// BlockFriendRequests reports whether incoming friend requests are disabled.
	BlockFriendRequests bool
	// BlockRoomInvites reports whether incoming room invitations are disabled.
	BlockRoomInvites bool
	// BlockFollowing reports whether friends may follow the player.
	BlockFollowing bool
}

// Record contains a player identity and profile pair.
type Record struct {
	// Player is the durable player identity.
	Player playermodel.Player

	// Profile is the durable player profile.
	Profile playermodel.Profile
}

// managerAssertion verifies Service implements Manager.
var managerAssertion Manager = (*Service)(nil)

// adminManagerAssertion verifies Service implements AdminManager.
var adminManagerAssertion AdminManager = (*Service)(nil)
