package model

import "testing"

// TestRotationValid verifies supported floor rotation values.
func TestRotationValid(t *testing.T) {
	valid := []Rotation{RotationNorth, RotationEast, RotationSouth, RotationWest}
	for _, rotation := range valid {
		if !rotation.Valid() {
			t.Fatalf("expected rotation %d to be valid", rotation)
		}
	}

	invalid := []Rotation{1, 3, 5, 7, -1, 8}
	for _, rotation := range invalid {
		if rotation.Valid() {
			t.Fatalf("expected rotation %d to be invalid", rotation)
		}
	}
}
