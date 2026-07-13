package live

import (
	"time"

	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
	playerservice "github.com/niflaot/pixels/internal/realm/player/service"
)

// Snapshot stores durable player data needed by runtime state.
type Snapshot struct {
	// ID identifies the player.
	ID int64

	// Username stores the visible player name.
	Username string

	// Look stores the avatar figure string.
	Look string

	// Gender stores the avatar gender code.
	Gender playermodel.Gender

	// Motto stores the public profile motto.
	Motto string

	// HomeRoomID stores the optional home room id.
	HomeRoomID *int64

	// AllowNameChange reports whether username changes are allowed.
	AllowNameChange bool

	// BubbleStyle stores the validated Nitro chat bubble style.
	BubbleStyle int32

	// BlockFriendRequests reports whether incoming friend requests are disabled.
	BlockFriendRequests bool

	// BlockRoomInvites reports whether incoming room invitations are disabled.
	BlockRoomInvites bool

	// BlockFollowing reports whether friends may follow the player to a room.
	BlockFollowing bool

	// Club contains the player's subscription entitlement.
	Club playermodel.Club

	// AllowTrade reports whether direct trading is enabled for the player.
	AllowTrade bool
}

// SnapshotFromRecord maps a persistent player record to a runtime snapshot.
func SnapshotFromRecord(record playerservice.Record) Snapshot {
	return Snapshot{
		ID:                  record.Player.ID,
		Username:            record.Player.Username,
		Look:                record.Profile.Look,
		Gender:              record.Profile.Gender,
		Motto:               record.Profile.Motto,
		HomeRoomID:          record.Profile.HomeRoomID,
		AllowNameChange:     record.Profile.AllowNameChange,
		BubbleStyle:         record.Profile.BubbleStyle,
		BlockFriendRequests: record.Profile.BlockFriendRequests,
		BlockRoomInvites:    record.Profile.BlockRoomInvites,
		BlockFollowing:      record.Profile.BlockFollowing,
		Club:                record.Player.Club,
		AllowTrade:          record.Player.AllowTrade,
	}
}

// ClubLevelAt returns the active club tier at one instant.
func (snapshot Snapshot) ClubLevelAt(now time.Time) playermodel.ClubLevel {
	return snapshot.Club.LevelAt(now)
}

// HasClubAt reports whether club access is active at one instant.
func (snapshot Snapshot) HasClubAt(now time.Time) bool {
	return snapshot.Club.ActiveAt(now)
}

// Valid reports whether the snapshot can create a live player.
func (snapshot Snapshot) Valid() bool {
	return snapshot.ID > 0 && snapshot.Username != ""
}
