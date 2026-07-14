// Package effect owns player effect inventory, activation, and room projection.
package effect

import "time"

const (
	// MaximumCharges caps one effect inventory stack.
	MaximumCharges int32 = 99
	// SourceAdmin identifies an administrative effect grant.
	SourceAdmin Source = "admin"
	// SourceCatalog identifies a catalog effect grant.
	SourceCatalog Source = "catalog"
	// SourceEffectGiver identifies a clicked effect-giver grant.
	SourceEffectGiver Source = "effect_giver"
	// SourceEffectTile identifies a walked-on effect-tile grant.
	SourceEffectTile Source = "effect_tile"
	// SourceRank identifies a synthetic permission-group effect.
	SourceRank Source = "rank"
)

// Source identifies how an effect was granted.
type Source string

// Effect contains one durable effect inventory entry.
type Effect struct {
	// PlayerID identifies the effect owner.
	PlayerID int64
	// ID identifies the Nitro effect.
	ID int32
	// DurationSeconds stores one charge duration; zero means permanent.
	DurationSeconds int32
	// ActivatedAt stores the current charge activation time.
	ActivatedAt *time.Time
	// RemainingCharges stores inventory charges.
	RemainingCharges int32
	// Synthetic reports whether the effect derives from rank configuration.
	Synthetic bool
}

// Permanent reports whether one charge has no expiry.
func (effect Effect) Permanent() bool {
	return effect.DurationSeconds == 0
}

// SecondsLeft returns the current active charge time remaining.
func (effect Effect) SecondsLeft(now time.Time) int32 {
	if effect.ActivatedAt == nil || effect.Permanent() {
		return 0
	}
	remaining := effect.DurationSeconds - int32(now.Sub(*effect.ActivatedAt)/time.Second)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// Expiration describes one charge consumed by the expiry scheduler.
type Expiration struct {
	// PlayerID identifies the affected player.
	PlayerID int64
	// EffectID identifies the expired effect.
	EffectID int32
	// RemainingCharges stores charges after expiration.
	RemainingCharges int32
	// Selected reports whether expiry cleared the room effect.
	Selected bool
}
