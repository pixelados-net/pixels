package handlers

import (
	"testing"
	"time"

	"github.com/niflaot/pixels/networking/codec"
)

// TestMinutesUntilWeekAtUsesTheNextMonday verifies Sunday and Monday boundaries.
func TestMinutesUntilWeekAtUsesTheNextMonday(t *testing.T) {
	tests := []struct {
		// now stores one UTC reference instant.
		now time.Time
		// want stores the expected whole-minute countdown.
		want int32
	}{
		{time.Date(2026, time.July, 19, 12, 0, 0, 0, time.UTC), 12 * 60},
		{time.Date(2026, time.July, 20, 0, 0, 0, 0, time.UTC), 7 * 24 * 60},
		{time.Date(2026, time.July, 21, 23, 30, 0, 0, time.UTC), 5*24*60 + 30},
	}
	for _, test := range tests {
		if got := minutesUntilWeekAt(test.now); got != test.want {
			t.Errorf("now=%s got=%d want=%d", test.now, got, test.want)
		}
	}
}

// TestDecodeNoopRejectsUnknownHeaders verifies unsupported packets fail closed.
func TestDecodeNoopRejectsUnknownHeaders(t *testing.T) {
	if err := decodeNoop(codec.Packet{Header: 65535}); err != codec.ErrUnexpectedHeader {
		t.Fatalf("error=%v", err)
	}
}
