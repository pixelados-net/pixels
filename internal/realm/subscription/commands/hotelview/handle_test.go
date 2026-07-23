package hotelview

import (
	"math"
	"testing"
	"time"
)

// TestSecondsUntil verifies future, past, and malformed countdowns.
func TestSecondsUntil(t *testing.T) {
	now := time.Date(2030, 4, 5, 12, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		name  string
		value string
		want  int64
	}{{name: "future", value: "2030-04-05 12:30", want: 1800}, {name: "past", value: "2030-04-05 11:00"}, {name: "invalid", value: "tomorrow"}} {
		t.Run(test.name, func(t *testing.T) {
			if got := secondsUntil(test.value, now); got != test.want {
				t.Fatalf("seconds=%d want=%d", got, test.want)
			}
		})
	}
}

// TestClampSeconds verifies protocol-range conversion.
func TestClampSeconds(t *testing.T) {
	if clampSeconds(-1) != 0 || clampSeconds(30) != 30 || clampSeconds(math.MaxInt64) != math.MaxInt32 {
		t.Fatal("unexpected countdown clamp")
	}
}

// BenchmarkSecondsUntil measures the pure countdown calculation.
func BenchmarkSecondsUntil(b *testing.B) {
	now := time.Date(2030, 4, 5, 12, 0, 0, 0, time.UTC)
	for b.Loop() {
		_ = secondsUntil("2030-04-05 12:30", now)
	}
}
