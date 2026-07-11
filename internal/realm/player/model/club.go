package model

import "time"

// ClubLevel identifies one Nitro club entitlement tier.
type ClubLevel int16

const (
	// ClubLevelNone identifies a player without club access.
	ClubLevelNone ClubLevel = iota
	// ClubLevelHC identifies Habbo Club access.
	ClubLevelHC
	// ClubLevelVIP identifies the highest club access tier.
	ClubLevelVIP
)

// Club contains one player's durable club entitlement.
type Club struct {
	// Level stores the granted club tier.
	Level ClubLevel
	// ExpiresAt stores the exclusive entitlement expiration time.
	ExpiresAt *time.Time
}

// LevelAt returns the active club level at one instant.
func (club Club) LevelAt(now time.Time) ClubLevel {
	if club.Level <= ClubLevelNone || club.ExpiresAt == nil || !now.Before(*club.ExpiresAt) {
		return ClubLevelNone
	}

	return club.Level
}

// ActiveAt reports whether club access is active at one instant.
func (club Club) ActiveAt(now time.Time) bool {
	return club.LevelAt(now) > ClubLevelNone
}
