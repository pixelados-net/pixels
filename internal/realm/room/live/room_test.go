package live

import (
	"errors"
	"testing"
)

// TestNewRoomRejectsInvalidSnapshot verifies snapshot validation.
func TestNewRoomRejectsInvalidSnapshot(t *testing.T) {
	_, err := NewRoom(Snapshot{})
	if !errors.Is(err, ErrInvalidRoom) {
		t.Fatalf("expected invalid room, got %v", err)
	}
}

// TestCanManageFurnitureAllowsOnlyOwner verifies the current owner-only room policy.
func TestCanManageFurnitureAllowsOnlyOwner(t *testing.T) {
	room, err := NewRoom(Snapshot{ID: 9, OwnerPlayerID: 7, MaxUsers: 1})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	if !room.CanManageFurniture(7) || room.CanManageFurniture(8) || room.CanManageFurniture(0) {
		t.Fatal("expected furniture management to be restricted to the room owner")
	}
}

// TestRoomJoinRejectsInvalidOccupant verifies occupant validation.
func TestRoomJoinRejectsInvalidOccupant(t *testing.T) {
	room, err := NewRoom(Snapshot{ID: 9, MaxUsers: 1})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}

	_, err = room.Join(Occupant{})
	if !errors.Is(err, ErrInvalidOccupant) {
		t.Fatalf("expected invalid occupant, got %v", err)
	}
}

// TestRoomIdleSinceTracksEmptyRoom verifies idle state after last leave.
func TestRoomIdleSinceTracksEmptyRoom(t *testing.T) {
	room, err := NewRoom(Snapshot{ID: 9, MaxUsers: 1})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	if _, err := room.Join(occupantForTest(7)); err != nil {
		t.Fatalf("join room: %v", err)
	}
	if idleSince := room.IdleSince(); idleSince != nil {
		t.Fatalf("expected active room, got idle at %v", idleSince)
	}

	if _, removed := room.Leave(7); !removed {
		t.Fatal("expected occupant removed")
	}
	if idleSince := room.IdleSince(); idleSince == nil {
		t.Fatal("expected idle room")
	}
}

// TestRoomSnapshotAndOccupants verifies active room snapshots.
func TestRoomSnapshotAndOccupants(t *testing.T) {
	room, err := NewRoom(Snapshot{ID: 9, MaxUsers: 2})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	if room.ID() != 9 || room.Snapshot().MaxUsers != 2 {
		t.Fatalf("unexpected snapshot %#v", room.Snapshot())
	}
	if _, err := room.Join(occupantForTest(7)); err != nil {
		t.Fatalf("join room: %v", err)
	}
	if len(room.Occupants()) != 1 || room.Occupancy().Count != 1 {
		t.Fatalf("unexpected occupants=%#v occupancy=%#v", room.Occupants(), room.Occupancy())
	}
}

// TestRoomCloseRejectsFutureJoins verifies closed room behavior.
func TestRoomCloseRejectsFutureJoins(t *testing.T) {
	room, err := NewRoom(Snapshot{ID: 9, MaxUsers: 1})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	room.Close()

	_, err = room.Join(occupantForTest(7))
	if !errors.Is(err, ErrRoomClosed) {
		t.Fatalf("expected room closed, got %v", err)
	}
}
