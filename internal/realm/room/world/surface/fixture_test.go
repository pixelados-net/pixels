package surface

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/internal/realm/room/world/grid"
)

// TestNewFixtureRejectsInvalidInput verifies fixture validation.
func TestNewFixtureRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		// name stores the test case name.
		name string

		// params stores fixture input.
		params FixtureParams
	}{
		{name: "top below z", params: FixtureParams{Point: grid.MustPoint(1, 1), Z: 2, Top: 1, State: StateOpen}},
		{name: "invalid state", params: FixtureParams{Point: grid.MustPoint(1, 1), Z: 1, Top: 1, State: StateInvalid}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewFixture(test.params)
			if !errors.Is(err, ErrInvalidFixture) {
				t.Fatalf("expected invalid fixture, got %v", err)
			}
		})
	}
}

// TestFixtureSectionDefaultsSource verifies generic fixture source behavior.
func TestFixtureSectionDefaultsSource(t *testing.T) {
	fixture := fixtureForTest(t, FixtureParams{
		Point:     grid.MustPoint(1, 1),
		Z:         3,
		Top:       5,
		Clearance: 2,
		State:     StateLay,
		Stacking:  true,
		SourceID:  99,
	})

	section := fixture.Section()
	if section.Source() != SourceFixture || section.Top() != 5 || section.Clearance() != 2 {
		t.Fatalf("unexpected section from fixture")
	}
	if !section.Stacking() || section.SourceID() != 99 || !section.Walkable() {
		t.Fatalf("unexpected fixture section metadata")
	}
}

// TestStateWalkable verifies movement state classification.
func TestStateWalkable(t *testing.T) {
	if StateInvalid.Walkable() || StateBlocked.Walkable() {
		t.Fatal("expected invalid and blocked states to reject movement")
	}
	if !StateOpen.Walkable() || !StateSit.Walkable() || !StateLay.Walkable() {
		t.Fatal("expected open, sit and lay states to accept movement")
	}
}
