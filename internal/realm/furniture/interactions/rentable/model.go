package rentable

import "time"

// State describes one rentable furniture instance.
type State struct {
	// ItemID identifies the furniture instance.
	ItemID int64
	// OwnerPlayerID identifies the permanent owner.
	OwnerPlayerID int64
	// RenterPlayerID identifies the current renter when active.
	RenterPlayerID *int64
	// ExpiresAt stores the current rental boundary.
	ExpiresAt *time.Time
	// PriceCredits stores one extension price.
	PriceCredits int32
}

// ActiveAt reports whether the rental is active.
func (state State) ActiveAt(now time.Time) bool {
	return state.RenterPlayerID != nil && state.ExpiresAt != nil && state.ExpiresAt.After(now)
}

// SecondsRemaining returns a bounded protocol duration.
func (state State) SecondsRemaining(now time.Time) int32 {
	if !state.ActiveAt(now) {
		return 0
	}
	seconds := int64(state.ExpiresAt.Sub(now) / time.Second)
	if seconds > int64(^uint32(0)>>1) {
		return int32(^uint32(0) >> 1)
	}
	return int32(seconds)
}
