package globalid

import "testing"

// TestParse verifies accepted active link formats.
func TestParse(t *testing.T) {
	for _, value := range []string{"130", "room:130", "ROOM:130"} {
		id, ok := parse(value)
		if !ok || id != 130 {
			t.Fatalf("parse %q = %d,%v", value, id, ok)
		}
	}
}

// TestParseRejectsUnknown verifies unrelated global identifiers are private failures.
func TestParseRejectsUnknown(t *testing.T) {
	for _, value := range []string{"", "group:130", "room:-1", "room:nope"} {
		if _, ok := parse(value); ok {
			t.Fatalf("expected %q rejection", value)
		}
	}
}
