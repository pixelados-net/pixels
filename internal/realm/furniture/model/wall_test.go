package model

import "testing"

// TestValidWallPosition verifies accepted and rejected Nitro wall coordinates.
func TestValidWallPosition(t *testing.T) {
	for _, test := range []struct {
		value string
		valid bool
	}{
		{value: ":w=2,3 l=4,5 r", valid: true},
		{value: ":w=0,0 l=0,0 l", valid: true},
		{value: ":w=-1,0 l=0,0 l", valid: false},
		{value: "2,3", valid: false},
	} {
		if ValidWallPosition(test.value) != test.valid {
			t.Fatalf("unexpected validity for %q", test.value)
		}
	}
}
