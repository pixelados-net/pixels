package rentable

import (
	"testing"
	"time"
)

// TestStateActiveAndRemaining verifies expiry is calculated rather than ticked.
func TestStateActiveAndRemaining(t *testing.T) {
	now := time.Unix(100, 0)
	renter := int64(7)
	expires := now.Add(90 * time.Second)
	state := State{RenterPlayerID: &renter, ExpiresAt: &expires}
	if !state.ActiveAt(now) || state.SecondsRemaining(now) != 90 {
		t.Fatalf("active=%t remaining=%d", state.ActiveAt(now), state.SecondsRemaining(now))
	}
	if state.ActiveAt(expires) || state.SecondsRemaining(expires) != 0 {
		t.Fatal("expiry boundary must be inactive")
	}
}

// BenchmarkSecondsRemaining measures the rental status hot calculation.
func BenchmarkSecondsRemaining(b *testing.B) {
	now := time.Unix(100, 0)
	renter := int64(7)
	expires := now.Add(time.Hour)
	state := State{RenterPlayerID: &renter, ExpiresAt: &expires}
	for range b.N {
		_ = state.SecondsRemaining(now)
	}
}
