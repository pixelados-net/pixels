package live

import (
	"errors"
	"testing"
)

// TestExitToDoorForcesPathAndMarksDeparture verifies kicked movement cannot be replaced.
func TestExitToDoorForcesPathAndMarksDeparture(t *testing.T) {
	room := unitAwayFromDoorForTest(t)
	walking, err := room.ExitToDoor(7)
	if err != nil || !walking {
		t.Fatalf("start exit walking=%v err=%v", walking, err)
	}
	if _, err := room.MoveTo(7, pointForTest(t, 1, 0)); !errors.Is(err, ErrUnitExiting) {
		t.Fatalf("expected forced path guard, got %v", err)
	}

	room.Tick()
	room.Tick()
	movements := room.Tick()
	if len(movements) != 1 || !movements[0].Settled || !movements[0].Exited || !movements[0].ForcedExit {
		t.Fatalf("expected settled door exit, got %#v", movements)
	}
}

// TestWalkingOntoDoorMarksDeparture verifies normal movement through the exit uses the same lifecycle.
func TestWalkingOntoDoorMarksDeparture(t *testing.T) {
	room := unitAwayFromDoorForTest(t)
	if _, err := room.MoveTo(7, pointForTest(t, 0, 0)); err != nil {
		t.Fatalf("walk to door: %v", err)
	}

	room.Tick()
	room.Tick()
	movements := room.Tick()
	if len(movements) != 1 || !movements[0].Exited || movements[0].ForcedExit {
		t.Fatalf("expected normal door exit, got %#v", movements)
	}
}

// unitAwayFromDoorForTest creates a settled unit two tiles away from the door.
func unitAwayFromDoorForTest(t *testing.T) *Room {
	t.Helper()
	room := worldRoomForTest(t, "000", 0, 0)
	if _, err := room.Join(occupantForTest(7)); err != nil {
		t.Fatalf("join room: %v", err)
	}
	if _, err := room.MoveTo(7, pointForTest(t, 2, 0)); err != nil {
		t.Fatalf("move away from door: %v", err)
	}
	room.Tick()
	room.Tick()
	room.Tick()

	return room
}
