// Package achievement owns durable player badges and respect.
package achievement

// Badge stores one durable player badge projection.
type Badge struct {
	// ID identifies the durable inventory entry.
	ID int64
	// Code identifies the badge asset.
	Code string
	// Equipped reports whether the player is wearing the badge.
	Equipped bool
	// Slot stores the active position or zero when unequipped.
	Slot int32
}
