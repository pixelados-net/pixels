package live

import (
	"errors"
	"testing"
	"time"

	roomdoorbell "github.com/niflaot/pixels/internal/realm/room/access/doorbell"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// TestNewRoomRejectsInvalidSnapshot verifies snapshot validation.
func TestNewRoomRejectsInvalidSnapshot(t *testing.T) {
	_, err := NewRoom(Snapshot{})
	if !errors.Is(err, ErrInvalidRoom) {
		t.Fatalf("expected invalid room, got %v", err)
	}
}

// TestRoomDoorbellRequiresOwnerAndDrainsAfterLeave verifies owner-bound waiting state.
func TestRoomDoorbellRequiresOwnerAndDrainsAfterLeave(t *testing.T) {
	room, err := NewRoom(Snapshot{ID: 9, OwnerPlayerID: 7, MaxUsers: 2})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	entry := roomdoorbell.Entry{PlayerID: 8, Username: "Guest", Handler: netconn.Context{ConnectionID: "guest", ConnectionKind: "websocket"}, RequestedAt: time.Now()}
	if room.RequestDoorbell(entry, false) {
		t.Fatal("expected request rejected without owner")
	}
	if _, err := room.Join(occupantForTest(7)); err != nil {
		t.Fatalf("join owner: %v", err)
	}
	if !room.RequestDoorbell(entry, true) || room.DoorbellLen() != 1 {
		t.Fatal("expected queued request")
	}
	if _, removed := room.Leave(7); !removed {
		t.Fatal("expected owner removed")
	}
	expired := room.DrainDoorbellWithoutApprover(false)
	if len(expired) != 1 || expired[0].Reason != roomdoorbell.ExpiredNoRightsHolder {
		t.Fatalf("unexpected drained requests %#v", expired)
	}
}

// TestRoomJoinWithCapacityAllowsExplicitBypass verifies capacity remains opt-in.
func TestRoomJoinWithCapacityAllowsExplicitBypass(t *testing.T) {
	room, err := NewRoom(Snapshot{ID: 9, MaxUsers: 1})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	if _, err := room.Join(occupantForTest(7)); err != nil {
		t.Fatalf("join first occupant: %v", err)
	}
	second := occupantForTest(8)
	second.ConnectionID = "second"
	if _, err := room.Join(second); !errors.Is(err, ErrRoomFull) {
		t.Fatalf("expected full room, got %v", err)
	}
	if _, err := room.JoinWithCapacity(second, true); err != nil {
		t.Fatalf("join with capacity bypass: %v", err)
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

// TestRoomCloseDrainsDoorbell verifies close releases waiting connection state.
func TestRoomCloseDrainsDoorbell(t *testing.T) {
	room, err := NewRoom(Snapshot{ID: 9, OwnerPlayerID: 7, MaxUsers: 2})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	if _, err := room.Join(occupantForTest(7)); err != nil {
		t.Fatalf("join owner: %v", err)
	}
	entry := roomdoorbell.Entry{PlayerID: 8, Username: "Guest", Handler: netconn.Context{ConnectionID: "guest", ConnectionKind: "websocket"}, RequestedAt: time.Now()}
	if !room.RequestDoorbell(entry, true) {
		t.Fatal("queue guest")
	}
	_, expired := room.CloseWithDoorbell()
	if len(expired) != 1 || expired[0].Reason != roomdoorbell.ExpiredRoomClosed || room.DoorbellLen() != 0 {
		t.Fatalf("unexpected close drain %#v", expired)
	}
}

// TestRoomRightsProjectionControlsFurniture verifies embedded runtime rights mutations.
func TestRoomRightsProjectionControlsFurniture(t *testing.T) {
	room, err := NewRoom(Snapshot{ID: 9, OwnerPlayerID: 1, MaxUsers: 2})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	room.ReplaceRights([]int64{2})
	if !room.CanManageFurniture(1) || !room.CanManageFurniture(2) || room.CanManageFurniture(3) {
		t.Fatalf("unexpected furniture rights owner=%v holder=%v guest=%v", room.CanManageFurniture(1), room.CanManageFurniture(2), room.CanManageFurniture(3))
	}
	room.RevokeRights(2)
	room.GrantRights(3)
	if room.HasRights(2) || !room.HasRights(3) {
		t.Fatalf("unexpected projected rights revoked=%v granted=%v", room.HasRights(2), room.HasRights(3))
	}
}

// TestRoomSettingsRuntimeStateIsAtomicAndEphemeral verifies active settings projection.
func TestRoomSettingsRuntimeStateIsAtomicAndEphemeral(t *testing.T) {
	room, err := NewRoom(Snapshot{ID: 9, OwnerPlayerID: 1, MaxUsers: 2})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	categoryID := int64(4)
	room.UpdateSettings(&categoryID, 8, 30, 1)
	room.SetMuteAll(true)
	snapshot := room.Snapshot()
	if snapshot.CategoryID == nil || *snapshot.CategoryID != 4 || snapshot.MaxUsers != 8 || !room.MuteAll() {
		t.Fatalf("unexpected active settings snapshot=%#v muted=%v", snapshot, room.MuteAll())
	}
	reloaded, err := NewRoom(snapshot)
	if err != nil {
		t.Fatalf("reload room: %v", err)
	}
	if reloaded.MuteAll() {
		t.Fatal("mute-all must not persist across runtime reload")
	}
}
