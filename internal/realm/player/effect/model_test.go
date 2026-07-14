package effect

import (
	"testing"
	"time"
)

// TestEffectTiming verifies permanent and active duration semantics.
func TestEffectTiming(t *testing.T) {
	now := time.Unix(100, 0)
	activated := now.Add(-10 * time.Second)
	tests := []struct {
		name      string
		effect    Effect
		permanent bool
		left      int32
	}{
		{name: "permanent", effect: Effect{}, permanent: true},
		{name: "inactive", effect: Effect{DurationSeconds: 60}, left: 0},
		{name: "active", effect: Effect{DurationSeconds: 60, ActivatedAt: &activated}, left: 50},
		{name: "expired", effect: Effect{DurationSeconds: 5, ActivatedAt: &activated}, left: 0},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.effect.Permanent() != test.permanent || test.effect.SecondsLeft(now) != test.left {
				t.Fatalf("unexpected timing for %#v", test.effect)
			}
		})
	}
}
