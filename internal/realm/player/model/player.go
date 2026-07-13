// Package model contains persistent player records.
package model

import (
	"time"

	"github.com/niflaot/pixels/internal/permission"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// Player contains durable player identity fields.
type Player struct {
	// Base contains shared durable record fields.
	sharedmodel.Base

	// Username is the unique visible player name.
	Username string

	// LastLoginAt is the last successful login time.
	LastLoginAt *time.Time

	// LastLogoutAt is the last recorded logout time.
	LastLogoutAt *time.Time

	// LastSeenAt is the last time the player was seen by profile systems.
	LastSeenAt *time.Time

	// Club contains the player's subscription entitlement.
	Club Club

	// AllowTrade reports whether the player may participate in direct trades.
	AllowTrade bool
}

// HolderID identifies the player permission holder.
func (player Player) HolderID() int64 {
	return player.ID
}

// HolderKind reports that Player is an individual permission holder.
func (player Player) HolderKind() permission.HolderKind {
	return permission.HolderPlayer
}
