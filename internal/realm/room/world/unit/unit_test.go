package unit

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	"github.com/niflaot/pixels/internal/realm/room/world/path"
	"github.com/niflaot/pixels/internal/realm/room/world/surface"
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

	step, moved, settled := roomUnit.Advance()
	if !moved || settled || step.Position != positionForTest(2, 1, 0) {
		t.Fatalf("unexpected first step %#v found=%v", step, ok)
	}
	if roomUnit.Previous() != positionForTest(1, 1, 0) || roomUnit.Position() != positionForTest(2, 1, 0) {
		t.Fatalf("unexpected moved position")
	}
	if roomUnit.BodyRotation() != RotationEast || roomUnit.HeadRotation() != RotationEast {
		t.Fatalf("expected east rotation")
	}
	assertStatus(t, roomUnit, StatusMove, "2,1,0")

	_, moved, settled = roomUnit.Advance()
	if !moved || settled || roomUnit.Moving() || roomUnit.PendingSteps() != 0 {
		t.Fatalf("expected completed movement")
	}
	if _, ok := roomUnit.Goal(); ok {
		t.Fatal("expected cleared goal")
	}
	assertStatus(t, roomUnit, StatusMove, "2,2,1")

	_, moved, settled = roomUnit.Advance()
	if moved || !settled {
		t.Fatalf("expected settled movement moved=%v settled=%v", moved, settled)
	}
	assertNoStatus(t, roomUnit, StatusMove)
}

// TestUnitSetPathClearsSitAndLay verifies standing up when a new goal is set.
func TestUnitSetPathClearsSitAndLay(t *testing.T) {
	roomUnit := unitForTest(t)
	roomUnit.SetStatus(StatusSit, "4")
	roomUnit.SetStatus(StatusLay, "4")

	roomUnit.SetPath(path.NewPath([]path.Step{{Position: positionForTest(2, 1, 0)}}))

	assertNoStatus(t, roomUnit, StatusSit)
	assertNoStatus(t, roomUnit, StatusLay)
}

// TestUnitSettleAppliesStatusAndForcedRotation verifies landed seat/lay state.
func TestUnitSettleAppliesStatusAndForcedRotation(t *testing.T) {
	roomUnit := unitForTest(t)
	roomUnit.SetPath(path.NewPath([]path.Step{{Position: positionForTest(2, 1, 0)}}))
	roomUnit.Advance()

	roomUnit.Settle(StatusSit, "4", RotationNorth, RotationWest)

	assertStatus(t, roomUnit, StatusSit, "4")
	assertNoStatus(t, roomUnit, StatusMove)
	if roomUnit.BodyRotation() != RotationNorth || roomUnit.HeadRotation() != RotationWest {
		t.Fatalf("expected forced rotation body=%d head=%d", roomUnit.BodyRotation(), roomUnit.HeadRotation())
	}
}

// TestUnitValidatePathDetectsStaleColumns verifies path staleness detection.
func TestUnitValidatePathDetectsStaleColumns(t *testing.T) {
	roomUnit := unitForTest(t)
	if err := roomUnit.ValidatePath(nil); err != nil {
		t.Fatalf("expected no validation without movement, got %v", err)
	}

	roomGrid, err := grid.Parse("000")
	if err != nil {
		t.Fatalf("parse grid: %v", err)
	}
	resolver, err := surface.NewResolver(roomGrid, nil)
	if err != nil {
		t.Fatalf("create resolver: %v", err)
	}
	finder := path.NewFinder(resolver, path.DefaultRules())
	roomPath, err := finder.Find(path.Position{Point: grid.MustPoint(0, 0), Z: 0}, grid.MustPoint(2, 0))
	if err != nil {
		t.Fatalf("find path: %v", err)
	}
	roomUnit.SetPath(roomPath)

	if err := roomUnit.ValidatePath(resolver); err != nil {
		t.Fatalf("expected fresh path to validate, got %v", err)
	}

	changed, err := surface.NewFixture(surface.FixtureParams{Point: grid.MustPoint(1, 0), Z: 1, Top: 1, State: surface.StateOpen})
	if err != nil {
		t.Fatalf("create fixture: %v", err)
	}
	if err := resolver.AddFixture(changed); err != nil {
		t.Fatalf("add fixture: %v", err)
	}

	if err := roomUnit.ValidatePath(resolver); !errors.Is(err, path.ErrInvalidPath) {
		t.Fatalf("expected invalid path after fixture change, got %v", err)
	}
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

// TestUnitSetHeightCorrectsVerticalPositionOnly verifies SetHeight updates Z without moving the
// unit's tile, used to fix a stale elevated Z after the furniture a unit stood on moves away.
func TestUnitSetHeightCorrectsVerticalPositionOnly(t *testing.T) {
	roomUnit := unitForTest(t)
	roomUnit.SetHeight(3)

	if roomUnit.Position() != positionForTest(1, 1, 3) {
		t.Fatalf("expected height corrected in place, got %#v", roomUnit.Position())
	}
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

// TestUnitFaceTowardRotatesBodyAndHead verifies direct facing behavior.
func TestUnitFaceTowardRotatesBodyAndHead(t *testing.T) {
	roomUnit := unitForTest(t)
	roomUnit.FaceToward(grid.MustPoint(2, 1))

	if roomUnit.BodyRotation() != RotationEast || roomUnit.HeadRotation() != RotationEast {
		t.Fatalf("expected east facing body=%d head=%d", roomUnit.BodyRotation(), roomUnit.HeadRotation())
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

	_, moved, settled := roomUnit.Advance()
	if moved || settled {
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
