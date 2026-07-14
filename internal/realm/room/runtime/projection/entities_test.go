package projection

import (
	"testing"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// TestUnitsAndStatuses verifies room projection records.
func TestUnitsAndStatuses(t *testing.T) {
	room := roomForTest(t)

	unitRecords := Units(room)
	statusRecords := Statuses(room)
	if len(unitRecords) != 1 || unitRecords[0].Name != "demo" || unitRecords[0].Gender != "M" {
		t.Fatalf("unexpected units %#v", unitRecords)
	}
	if len(statusRecords) != 1 || statusRecords[0].RoomIndex != 1 {
		t.Fatalf("unexpected statuses %#v", statusRecords)
	}

	filtered := Units(room, 99)
	if len(filtered) != 0 {
		t.Fatalf("expected filtered units, got %#v", filtered)
	}
}

// TestStatusesSnapshotUsesCurrentPosition verifies the room-enter snapshot anchors each unit at its
// current tile, not the tile it stepped from, so newcomers never see occupants one step behind.
func TestStatusesSnapshotUsesCurrentPosition(t *testing.T) {
	room, err := roomlive.NewRoom(roomlive.Snapshot{ID: 9, MaxUsers: 5})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	roomGrid, err := grid.Parse("00", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatalf("parse grid: %v", err)
	}
	if err := room.LoadWorld(roomlive.WorldConfig{
		Grid: roomGrid,
		Door: worldpath.Position{Point: grid.MustPoint(0, 0)},
		Body: worldunit.RotationEast,
		Head: worldunit.RotationEast,
	}); err != nil {
		t.Fatalf("load world: %v", err)
	}
	if _, err := room.Join(roomlive.Occupant{
		PlayerID: 7, Username: "demo", ConnectionID: "conn", ConnectionKind: "websocket",
	}); err != nil {
		t.Fatalf("join room: %v", err)
	}
	if _, err := room.MoveTo(7, grid.MustPoint(1, 0)); err != nil {
		t.Fatalf("move unit: %v", err)
	}
	if movements := room.Tick(); len(movements) != 1 {
		t.Fatalf("expected one movement, got %#v", movements)
	}

	statusRecords := Statuses(room)
	if len(statusRecords) != 1 || statusRecords[0].X != 1 || statusRecords[0].Y != 0 {
		t.Fatalf("expected snapshot at current tile (1,0), got %#v", statusRecords)
	}
}

// TestNilRoomProjection verifies nil room guards.
func TestNilRoomProjection(t *testing.T) {
	if Units(nil) != nil || Statuses(nil) != nil {
		t.Fatal("expected nil projections")
	}
}

// TestStatusActionsKeepDanceOnItsDedicatedPacket verifies protocol separation.
func TestStatusActionsKeepDanceOnItsDedicatedPacket(t *testing.T) {
	actions := statusActions([]worldunit.Status{{Key: worldunit.StatusDance, Value: "3"}, {Key: worldunit.StatusSign, Value: "7"}})
	if len(actions) != 1 || actions[0].Key != worldunit.StatusSign {
		t.Fatalf("unexpected actions %#v", actions)
	}
}

// TestMovementStatuses verifies movement projection records.
func TestMovementStatuses(t *testing.T) {
	movements := []roomlive.Movement{{
		Unit: roomlive.UnitSnapshot{
			UnitID:   1,
			Previous: worldpath.Position{Point: grid.MustPoint(1, 3), Z: 0},
			Moving:   true,
		},
		Step:  worldpath.Step{Position: worldpath.Position{Point: grid.MustPoint(2, 3), Z: 1}},
		Moved: true,
	}}

	statusRecords := MovementStatuses(movements)
	if len(statusRecords) != 1 || statusRecords[0].Actions[0].Value != "2,3,1" || statusRecords[0].X != 1 {
		t.Fatalf("unexpected movement statuses %#v", statusRecords)
	}
}

// TestMovementStatusesStopsAtFinalPosition verifies the final movement status.
func TestMovementStatusesStopsAtFinalPosition(t *testing.T) {
	movements := []roomlive.Movement{{
		Unit: roomlive.UnitSnapshot{
			UnitID:   1,
			Position: worldpath.Position{Point: grid.MustPoint(2, 3), Z: 1},
			Previous: worldpath.Position{Point: grid.MustPoint(1, 3), Z: 0},
			Moving:   false,
		},
		Step:    worldpath.Step{Position: worldpath.Position{Point: grid.MustPoint(2, 3), Z: 1}},
		Settled: true,
	}}

	statusRecords := MovementStatuses(movements)
	if len(statusRecords) != 1 || len(statusRecords[0].Actions) != 0 || statusRecords[0].X != 2 || statusRecords[0].Y != 3 {
		t.Fatalf("unexpected final movement statuses %#v", statusRecords)
	}
}

// roomForTest creates a projected live room.
func roomForTest(t *testing.T) *roomlive.Room {
	t.Helper()

	room, err := roomlive.NewRoom(roomlive.Snapshot{ID: 9, MaxUsers: 5})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	roomGrid, err := grid.Parse("0", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatalf("parse grid: %v", err)
	}
	err = room.LoadWorld(roomlive.WorldConfig{
		Grid: roomGrid,
		Door: worldpath.Position{Point: grid.MustPoint(0, 0)},
		Body: worldunit.RotationEast,
		Head: worldunit.RotationEast,
	})
	if err != nil {
		t.Fatalf("load world: %v", err)
	}
	if _, err := room.Join(roomlive.Occupant{
		PlayerID: 7, Username: "demo", Motto: "hi", Figure: "hd-180-1",
		ConnectionID: "conn", ConnectionKind: "websocket",
	}); err != nil {
		t.Fatalf("join room: %v", err)
	}

	return room
}
