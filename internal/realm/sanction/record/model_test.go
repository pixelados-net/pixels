package record

import (
	"testing"
	"time"
)

// TestKindsAndActiveState verifies supported kinds and timestamp-only activity.
func TestKindsAndActiveState(t *testing.T) {
	for _, kind := range []Kind{KindBan, KindMute, KindWarn, KindTradeLock, KindKick} {
		if !kind.Valid() {
			t.Fatalf("kind=%q", kind)
		}
	}
	if Kind("unknown").Valid() || !KindWarn.Instant() || !KindKick.Instant() || KindMute.Instant() {
		t.Fatal("kind classification mismatch")
	}
	now := time.Now()
	expires := now.Add(time.Minute)
	value := Punishment{Kind: KindMute, ExpiresAt: &expires}
	if !value.ActiveAt(now) || value.ActiveAt(expires) {
		t.Fatalf("activity mismatch value=%+v", value)
	}
	value.Kind = KindWarn
	if value.ActiveAt(now) {
		t.Fatal("instant punishment became active")
	}
}
