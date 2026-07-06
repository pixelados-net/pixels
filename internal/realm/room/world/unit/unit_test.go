package unit

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	"github.com/niflaot/pixels/internal/realm/room/world/path"
)

// TestNewRejectsInvalidUnit verifies unit creation validation.
func TestNewRejectsInvalidUnit(t *testing.T) {
	_, err := New(Params{})
	if !errors.Is(err, ErrInvalidUnit) {
		t.Fatalf("expected invalid unit, got %v", err)
	}
}

// TestUnitExposesIdentityAndPosition verifies unit creation state.
func TestUnitExposesIdentityAndPosition(t *testing.T) {
	roomUnit := unitForTest(t)

	if roomUnit.ID() != 1 || roomUnit.OwnerID() != 77 || roomUnit.Kind() != KindPlayer {
		t.Fatalf("unexpected unit identity")
	}
	if roomUnit.Position() != positionForTest(1, 1, 0) || roomUnit.Previous() != positionForTest(1, 1, 0) {
		t.Fatalf("unexpected unit position")
	}
	if roomUnit.BodyRotation() != RotationSouth || roomUnit.HeadRotation() != RotationSouth {
		t.Fatalf("unexpected unit rotation")
	}
}

// TestUnitAdvancesPath verifies pending movement state.
func TestUnitAdvancesPath(t *testing.T) {
	roomUnit := unitForTest(t)
	roomUnit.SetPath(path.NewPath([]path.Step{
		{Position: positionForTest(2, 1, 0)},
		{Position: positionForTest(2, 2, 1), Diagonal: false},
	}))

	goal, ok := roomUnit.Goal()
	if !ok || goal != positionForTest(2, 2, 1) {
		t.Fatalf("unexpected goal %#v found=%v", goal, ok)
	}
	if !roomUnit.Moving() || roomUnit.PendingSteps() != 2 {
		t.Fatalf("expected pending movement")
	}
	assertStatus(t, roomUnit, StatusMove, "2,1,0")

	step, ok := roomUnit.Advance()
	if !ok || step.Position != positionForTest(2, 1, 0) {
		t.Fatalf("unexpected first step %#v found=%v", step, ok)
	}
	if roomUnit.Previous() != positionForTest(1, 1, 0) || roomUnit.Position() != positionForTest(2, 1, 0) {
		t.Fatalf("unexpected moved position")
	}
	if roomUnit.BodyRotation() != RotationEast || roomUnit.HeadRotation() != RotationEast {
		t.Fatalf("expected east rotation")
	}
	assertStatus(t, roomUnit, StatusMove, "2,2,1")

	_, ok = roomUnit.Advance()
	if !ok || roomUnit.Moving() || roomUnit.PendingSteps() != 0 {
		t.Fatalf("expected completed movement")
	}
	if _, ok := roomUnit.Goal(); ok {
		t.Fatal("expected cleared goal")
	}
	assertNoStatus(t, roomUnit, StatusMove)
}

// TestUnitClearPath verifies movement cancellation.
func TestUnitClearPath(t *testing.T) {
	roomUnit := unitForTest(t)
	roomUnit.SetPath(path.NewPath([]path.Step{{Position: positionForTest(2, 1, 0)}}))
	roomUnit.ClearPath()

	if roomUnit.Moving() || roomUnit.PendingSteps() != 0 {
		t.Fatal("expected cleared movement")
	}
	assertNoStatus(t, roomUnit, StatusMove)
}

// TestUnitStatusesAreStable verifies status mutation and ordering.
func TestUnitStatusesAreStable(t *testing.T) {
	roomUnit := unitForTest(t)
	roomUnit.SetStatus(StatusSit, "1.0")
	roomUnit.SetStatus(StatusLay, "0.5")
	roomUnit.ClearStatus(StatusLay)

	statuses := roomUnit.Statuses()
	if len(statuses) != 1 || statuses[0].Key != StatusSit || statuses[0].Value != "1.0" {
		t.Fatalf("unexpected statuses %#v", statuses)
	}
}

// TestRotationBetweenDirections verifies directional rotation mapping.
func TestRotationBetweenDirections(t *testing.T) {
	tests := []struct {
		// name stores the test case name.
		name string

		// toX stores the target x coordinate.
		toX uint16

		// toY stores the target y coordinate.
		toY uint16

		// expected stores the expected rotation.
		expected Rotation
	}{
		{name: "north", toX: 1, toY: 0, expected: RotationNorth},
		{name: "north east", toX: 2, toY: 0, expected: RotationNorthEast},
		{name: "east", toX: 2, toY: 1, expected: RotationEast},
		{name: "south east", toX: 2, toY: 2, expected: RotationSouthEast},
		{name: "south", toX: 1, toY: 2, expected: RotationSouth},
		{name: "south west", toX: 0, toY: 2, expected: RotationSouthWest},
		{name: "west", toX: 0, toY: 1, expected: RotationWest},
		{name: "north west", toX: 0, toY: 0, expected: RotationNorthWest},
		{name: "same", toX: 1, toY: 1, expected: RotationSouth},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rotation := RotationBetween(1, 1, test.toX, test.toY)
			if rotation != test.expected {
				t.Fatalf("expected rotation %d, got %d", test.expected, rotation)
			}
		})
	}
}

// TestUnitAdvanceWithoutPathReportsMissingStep verifies empty movement.
func TestUnitAdvanceWithoutPathReportsMissingStep(t *testing.T) {
	roomUnit := unitForTest(t)

	_, ok := roomUnit.Advance()
	if ok {
		t.Fatal("expected missing step")
	}
}

// unitForTest creates a room unit for tests.
func unitForTest(t *testing.T) *Unit {
	t.Helper()

	roomUnit, err := New(Params{
		ID:       1,
		OwnerID:  77,
		Kind:     KindPlayer,
		Position: positionForTest(1, 1, 0),
		Body:     RotationSouth,
		Head:     RotationSouth,
	})
	if err != nil {
		t.Fatalf("create unit: %v", err)
	}

	return roomUnit
}

// positionForTest creates a path position for tests.
func positionForTest(x int, y int, z grid.Height) path.Position {
	return path.Position{Point: grid.MustPoint(x, y), Z: z}
}

// assertStatus verifies a unit status.
func assertStatus(t *testing.T, roomUnit *Unit, key string, value string) {
	t.Helper()

	for _, status := range roomUnit.Statuses() {
		if status.Key == key && status.Value == value {
			return
		}
	}
	t.Fatalf("expected status %s=%s in %#v", key, value, roomUnit.Statuses())
}

// assertNoStatus verifies a missing unit status.
func assertNoStatus(t *testing.T, roomUnit *Unit, key string) {
	t.Helper()

	for _, status := range roomUnit.Statuses() {
		if status.Key == key {
			t.Fatalf("unexpected status %s in %#v", key, roomUnit.Statuses())
		}
	}
}
