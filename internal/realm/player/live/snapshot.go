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

	// CreatedAt stores the durable account registration instant.
	CreatedAt time.Time

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

	// VolumeSystem stores the client system volume.
	VolumeSystem int32

	// VolumeFurniture stores the client furniture volume.
	VolumeFurniture int32

	// VolumeTrax stores the client music volume.
	VolumeTrax int32

	// OldChat reports whether the legacy chat renderer is selected.
	OldChat bool

	// CameraFollowBlocked reports whether automatic camera following is disabled.
	CameraFollowBlocked bool

	// SafetyLocked reports the hotel-managed safety state.
	SafetyLocked bool

	// RespectsReceived stores the durable user respect total.
	RespectsReceived int32

	// RespectsRemaining stores the remaining daily user respect allowance.
	RespectsRemaining int32

	// RespectsPetRemaining stores the remaining daily pet respect allowance.
	RespectsPetRemaining int32

	// LastAccessDate stores the previous login date expected by Nitro.
	LastAccessDate string

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

	// ActiveEffectID stores the selected avatar effect.
	ActiveEffectID *int32

	// AchievementScore stores durable progression score.
	AchievementScore int32

	// Sanctions stores active global punishment projection.
	Sanctions Sanctions
}

// Sanctions stores allocation-free active global punishment state.
type Sanctions struct {
	// MuteUntil stores finite mute expiry.
	MuteUntil time.Time
	// MutePermanent reports an active permanent mute.
	MutePermanent bool
	// TradeLockUntil stores finite trade-lock expiry.
	TradeLockUntil time.Time
	// TradeLockPermanent reports an active permanent trade lock.
	TradeLockPermanent bool
}

// MutedAt reports whether global chat is muted at one instant.
func (sanctions Sanctions) MutedAt(now time.Time) bool {
	return sanctions.MutePermanent || sanctions.MuteUntil.After(now)
}

// TradeLockedAt reports whether direct trading is globally locked at one instant.
func (sanctions Sanctions) TradeLockedAt(now time.Time) bool {
	return sanctions.TradeLockPermanent || sanctions.TradeLockUntil.After(now)
}

// SetSanctions replaces the live global punishment projection.
func (player *Player) SetSanctions(sanctions Sanctions) {
	player.mutex.Lock()
	defer player.mutex.Unlock()
	player.snapshot.Sanctions = sanctions
}

// SetClub replaces the live club entitlement projection.
func (player *Player) SetClub(club playermodel.Club) {
	player.mutex.Lock()
	defer player.mutex.Unlock()
	player.snapshot.Club = club
}

// SetAllowTrade replaces the live direct-trade eligibility projection.
func (player *Player) SetAllowTrade(allow bool) {
	player.mutex.Lock()
	defer player.mutex.Unlock()
	player.snapshot.AllowTrade = allow
}

// SetActiveEffect replaces the live selected avatar effect.
func (player *Player) SetActiveEffect(effectID *int32) {
	player.mutex.Lock()
	defer player.mutex.Unlock()
	player.snapshot.ActiveEffectID = effectID
}

// SetClientSettings replaces settings projected into the live player.
func (player *Player) SetClientSettings(system int32, furniture int32, trax int32, oldChat bool, cameraFollowBlocked bool, safetyLocked bool) {
	player.mutex.Lock()
	defer player.mutex.Unlock()
	player.snapshot.VolumeSystem = system
	player.snapshot.VolumeFurniture = furniture
	player.snapshot.VolumeTrax = trax
	player.snapshot.OldChat = oldChat
	player.snapshot.CameraFollowBlocked = cameraFollowBlocked
	player.snapshot.SafetyLocked = safetyLocked
}

// SetHomeRoom replaces the live home-room projection.
func (player *Player) SetHomeRoom(roomID *int64) {
	player.mutex.Lock()
	defer player.mutex.Unlock()
	player.snapshot.HomeRoomID = roomID
}

// SetProfile replaces live avatar presentation fields.
func (player *Player) SetProfile(figure string, gender playermodel.Gender, motto string) {
	player.mutex.Lock()
	defer player.mutex.Unlock()
	player.snapshot.Look = figure
	player.snapshot.Gender = gender
	player.snapshot.Motto = motto
}

// SetRespect replaces live respect counters.
func (player *Player) SetRespect(received int32, remaining int32, petRemaining int32) {
	player.mutex.Lock()
	defer player.mutex.Unlock()
	player.snapshot.RespectsReceived = received
	player.snapshot.RespectsRemaining = remaining
	player.snapshot.RespectsPetRemaining = petRemaining
}

// SetAchievementScore replaces the live durable progression score.
func (player *Player) SetAchievementScore(score int32) {
	player.mutex.Lock()
	defer player.mutex.Unlock()
	player.snapshot.AchievementScore = score
}

// SetUsername replaces the live visible identity and one-shot policy.
func (player *Player) SetUsername(username string, allowChange bool) {
	player.mutex.Lock()
	defer player.mutex.Unlock()
	player.snapshot.Username = username
	player.snapshot.AllowNameChange = allowChange
}

// SnapshotFromRecord maps a persistent player record to a runtime snapshot.
func SnapshotFromRecord(record playerservice.Record) Snapshot {
	snapshot := Snapshot{
		ID:                  record.Player.ID,
		CreatedAt:           record.Player.CreatedAt,
		Username:            record.Player.Username,
		Look:                record.Profile.Look,
		Gender:              record.Profile.Gender,
		Motto:               record.Profile.Motto,
		HomeRoomID:          record.Profile.HomeRoomID,
		AllowNameChange:     record.Profile.AllowNameChange,
		BubbleStyle:         record.Profile.BubbleStyle,
		VolumeSystem:        100,
		VolumeFurniture:     100,
		VolumeTrax:          100,
		BlockFriendRequests: record.Profile.BlockFriendRequests,
		BlockRoomInvites:    record.Profile.BlockRoomInvites,
		BlockFollowing:      record.Profile.BlockFollowing,
		Club:                record.Player.Club,
		AllowTrade:          record.Player.AllowTrade,
		ActiveEffectID:      record.Player.ActiveEffectID,
		AchievementScore:    record.Player.AchievementScore,
	}
	if record.Player.LastLoginAt != nil {
		snapshot.LastAccessDate = record.Player.LastLoginAt.Format("02-01-2006 15:04:05")
	}
	return snapshot
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
