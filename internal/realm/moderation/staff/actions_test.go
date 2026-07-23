package staff

import "testing"

// TestBanDurationHours verifies every ban option emitted by the current Nitro UI.
func TestBanDurationHours(t *testing.T) {
	tests := []struct {
		index int32
		hours int32
		valid bool
	}{{2, 18, true}, {3, 168, true}, {4, 720, true}, {5, 720, true}, {6, 0, true}, {7, 0, true}, {1, 0, false}}
	for _, test := range tests {
		hours, valid := banDurationHours(test.index)
		if hours != test.hours || valid != test.valid {
			t.Fatalf("index=%d hours=%d valid=%v", test.index, hours, valid)
		}
	}
}

// TestTradeLockDurationHours verifies finite and permanent Nitro durations.
func TestTradeLockDurationHours(t *testing.T) {
	tests := []struct {
		minutes int32
		hours   int32
		valid   bool
	}{{10080, 168, true}, {permanentTradeLockMinutes, 0, true}, {61, 2, true}, {0, 0, false}}
	for _, test := range tests {
		hours, valid := tradeLockDurationHours(test.minutes)
		if hours != test.hours || valid != test.valid {
			t.Fatalf("minutes=%d hours=%d valid=%v", test.minutes, hours, valid)
		}
	}
}
