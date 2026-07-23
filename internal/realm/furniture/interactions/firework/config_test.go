package firework

import (
	"testing"
	"time"
)

// TestConfigNormalizeBoundsRecharge verifies invalid durations use the safe fallback.
func TestConfigNormalizeBoundsRecharge(t *testing.T) {
	for _, value := range []time.Duration{0, -time.Second, 6 * time.Minute} {
		if got := (Config{DefaultRecharge: value}).Normalize().DefaultRecharge; got != defaultRecharge {
			t.Fatalf("value=%s got=%s", value, got)
		}
	}
	if got := (Config{DefaultRecharge: time.Second}).Normalize().DefaultRecharge; got != time.Second {
		t.Fatalf("got=%s", got)
	}
}

// BenchmarkRecharge parses a representative furniture override.
func BenchmarkRecharge(b *testing.B) {
	service := New(Config{}, nil, nil, nil)
	for range b.N {
		_ = service.recharge("15")
	}
}
