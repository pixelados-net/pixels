package unit

import (
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	"github.com/niflaot/pixels/internal/realm/room/world/path"
)

// TestPostureMovementAndIdleClearTransientState verifies avatar action invariants.
func TestPostureMovementAndIdleClearTransientState(t *testing.T) {
	roomUnit := unitForTest(t)
	roomUnit.SetStatus(StatusDance, "3")
	roomUnit.SetFloorPosture(true)
	if roomUnit.HasStatus(StatusDance) || !roomUnit.HasStatus(StatusSit) {
		t.Fatalf("unexpected posture statuses %#v", roomUnit.Statuses())
	}
	roomUnit.SetStatus(StatusDance, "2")
	roomUnit.SetPath(path.NewPath([]path.Step{{Position: path.Position{Point: grid.MustPoint(2, 1)}}}))
	if roomUnit.HasStatus(StatusDance) || roomUnit.HasStatus(StatusSit) {
		t.Fatalf("unexpected movement statuses %#v", roomUnit.Statuses())
	}
	now := time.Unix(100, 0)
	roomUnit.SetIdleAt(true, now)
	if !roomUnit.Idle() || !roomUnit.IdleSince().Equal(now) {
		t.Fatalf("unexpected idle projection idle=%t since=%s", roomUnit.Idle(), roomUnit.IdleSince())
	}
	roomUnit.SetManualIdleAt(true, now)
	if !roomUnit.Idle() || !roomUnit.ManualIdle() {
		t.Fatalf("unexpected manual idle projection idle=%t manual=%t", roomUnit.Idle(), roomUnit.ManualIdle())
	}
	roomUnit.SetIdleAt(false, now.Add(time.Second))
	if roomUnit.Idle() || roomUnit.ManualIdle() || !roomUnit.IdleSince().IsZero() {
		t.Fatalf("unexpected cleared idle projection idle=%t since=%s", roomUnit.Idle(), roomUnit.IdleSince())
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
