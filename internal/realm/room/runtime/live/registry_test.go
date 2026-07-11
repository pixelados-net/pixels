package live

import (
	"context"
	"errors"
	"testing"
	"time"

	roomdoorbell "github.com/niflaot/pixels/internal/realm/room/access/doorbell"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// TestRegistryJoinLeavePublishesOccupancy verifies active room occupancy flow.
func TestRegistryJoinLeavePublishesOccupancy(t *testing.T) {
	var published []Occupancy
	registry := NewRegistry(func(_ context.Context, occupancy Occupancy) error {
		published = append(published, occupancy)
		return nil
	})

	if _, err := registry.Activate(Snapshot{ID: 9, MaxUsers: 2}); err != nil {
		t.Fatalf("activate room: %v", err)
	}

	occupancy, err := registry.Join(context.Background(), 9, occupantForTest(7))
	if err != nil {
		t.Fatalf("join room: %v", err)
	}
	if occupancy.Count != 1 || len(published) != 1 {
		t.Fatalf("unexpected occupancy=%#v published=%#v", occupancy, published)
	}

	occupancy, removed, err := registry.Leave(context.Background(), 7)
	if err != nil {
		t.Fatalf("leave room: %v", err)
	}
	if !removed || occupancy.Count != 0 || len(published) != 2 {
		t.Fatalf("unexpected leave occupancy=%#v removed=%v published=%#v", occupancy, removed, published)
	}
}

// TestRegistryJoinRejectsMissingRoom verifies rooms must be active before joins.
func TestRegistryJoinRejectsMissingRoom(t *testing.T) {
	_, err := NewRegistry(nil).Join(context.Background(), 9, occupantForTest(7))
	if !errors.Is(err, ErrRoomNotFound) {
		t.Fatalf("expected room not found, got %v", err)
	}
}

// TestRegistryJoinRejectsFullRoom verifies room capacity.
func TestRegistryJoinRejectsFullRoom(t *testing.T) {
	registry := NewRegistry(nil)
	if _, err := registry.Activate(Snapshot{ID: 9, MaxUsers: 1}); err != nil {
		t.Fatalf("activate room: %v", err)
	}
	if _, err := registry.Join(context.Background(), 9, occupantForTest(7)); err != nil {
		t.Fatalf("join first occupant: %v", err)
	}

	_, err := registry.Join(context.Background(), 9, occupantForTest(8))
	if !errors.Is(err, ErrRoomFull) {
		t.Fatalf("expected room full, got %v", err)
	}
}

// TestRegistryJoinSameRoomPreservesUnit verifies rejoin does not recreate units.
func TestRegistryJoinSameRoomPreservesUnit(t *testing.T) {
	registry := NewRegistry(nil)
	active, err := registry.Activate(Snapshot{ID: 9, MaxUsers: 2})
	if err != nil {
		t.Fatalf("activate room: %v", err)
	}
	if err := active.LoadWorld(worldConfigForTest(t)); err != nil {
		t.Fatalf("load world: %v", err)
	}
	if _, err := registry.Join(context.Background(), 9, occupantForTest(7)); err != nil {
		t.Fatalf("join room: %v", err)
	}
	first := active.Units()[0].UnitID

	if _, err := registry.Join(context.Background(), 9, occupantForTest(7)); err != nil {
		t.Fatalf("rejoin room: %v", err)
	}
	units := active.Units()
	if len(units) != 1 || units[0].UnitID != first {
		t.Fatalf("unexpected units %#v", units)
	}
}

// TestRegistryCloseUnregistersRoom verifies close behavior.
func TestRegistryCloseUnregistersRoom(t *testing.T) {
	registry := NewRegistry(nil)
	if _, err := registry.Activate(Snapshot{ID: 9, MaxUsers: 1}); err != nil {
		t.Fatalf("activate room: %v", err)
	}
	if _, err := registry.Join(context.Background(), 9, occupantForTest(7)); err != nil {
		t.Fatalf("join room: %v", err)
	}

	occupancy, closed, err := registry.Close(context.Background(), 9)
	if err != nil {
		t.Fatalf("close room: %v", err)
	}
	if !closed || occupancy.Count != 0 || registry.Count() != 0 {
		t.Fatalf("unexpected close occupancy=%#v closed=%v count=%d", occupancy, closed, registry.Count())
	}
}

// TestRegistryActivateReturnsExistingRoom verifies active room reuse.
func TestRegistryActivateReturnsExistingRoom(t *testing.T) {
	registry := NewRegistry(nil)
	first, err := registry.Activate(Snapshot{ID: 9, MaxUsers: 1})
	if err != nil {
		t.Fatalf("activate room: %v", err)
	}
	second, err := registry.Activate(Snapshot{ID: 9, MaxUsers: 1})
	if err != nil {
		t.Fatalf("activate existing room: %v", err)
	}

	if first != second || len(registry.Snapshot()) != 1 {
		t.Fatalf("expected reused room")
	}
}

// TestRegistryRemovePlayerReportsMissing verifies missing player cleanup.
func TestRegistryRemovePlayerReportsMissing(t *testing.T) {
	_, removed, err := NewRegistry(nil).RemovePlayer(context.Background(), 7)
	if err != nil {
		t.Fatalf("remove player: %v", err)
	}
	if removed {
		t.Fatal("expected missing player")
	}
}

// TestRegistryUnloadIdleClosesEmptyRooms verifies idle room cleanup.
func TestRegistryUnloadIdleClosesEmptyRooms(t *testing.T) {
	registry := NewRegistry(nil)
	if _, err := registry.Activate(Snapshot{ID: 9, MaxUsers: 1}); err != nil {
		t.Fatalf("activate room: %v", err)
	}
	if _, err := registry.Join(context.Background(), 9, occupantForTest(7)); err != nil {
		t.Fatalf("join room: %v", err)
	}
	if _, removed, err := registry.Leave(context.Background(), 7); err != nil || !removed {
		t.Fatalf("leave room removed=%v err=%v", removed, err)
	}

	closed, err := registry.UnloadIdle(context.Background(), time.Second, time.Now().Add(time.Hour))
	if err != nil {
		t.Fatalf("unload idle: %v", err)
	}
	if len(closed) != 1 || registry.Count() != 0 {
		t.Fatalf("unexpected closed=%#v count=%d", closed, registry.Count())
	}
}

// TestRegistryKeepsDoorbellWhileApproverRemains verifies authorized responder draining.
func TestRegistryKeepsDoorbellWhileApproverRemains(t *testing.T) {
	var expired []roomdoorbell.Expired
	registry := NewRegistry(nil,
		WithDoorbellApprover(func(_ context.Context, room *Room) (bool, error) {
			return room.Occupancy().Count > 0, nil
		}),
		WithDoorbellPublisher(func(_ context.Context, _ *Room, entries []roomdoorbell.Expired) error {
			expired = append(expired, entries...)
			return nil
		}),
	)
	active, err := registry.Activate(Snapshot{ID: 9, OwnerPlayerID: 7, MaxUsers: 3})
	if err != nil {
		t.Fatalf("activate room: %v", err)
	}
	if _, err := registry.Join(context.Background(), 9, occupantForTest(7)); err != nil {
		t.Fatalf("join owner: %v", err)
	}
	if _, err := registry.Join(context.Background(), 9, occupantForTest(2)); err != nil {
		t.Fatalf("join moderator: %v", err)
	}
	entry := roomdoorbell.Entry{PlayerID: 8, Username: "Guest", Handler: netconn.Context{ConnectionID: "guest", ConnectionKind: "websocket"}, RequestedAt: time.Now()}
	if !active.RequestDoorbell(entry, true) {
		t.Fatal("queue doorbell")
	}
	if _, _, err := registry.Leave(context.Background(), 7); err != nil || active.DoorbellLen() != 1 {
		t.Fatalf("owner leave drained queue len=%d err=%v", active.DoorbellLen(), err)
	}
	if _, _, err := registry.Leave(context.Background(), 2); err != nil || active.DoorbellLen() != 0 || len(expired) != 1 {
		t.Fatalf("moderator leave queue=%d expired=%#v err=%v", active.DoorbellLen(), expired, err)
	}
}

// occupantForTest creates an active room occupant.
func occupantForTest(playerID int64) Occupant {
	return Occupant{PlayerID: playerID, Username: "demo", ConnectionID: netconn.ID("conn"), ConnectionKind: netconn.Kind("websocket")}
}

// worldConfigForTest creates a flat world config.
func worldConfigForTest(t testing.TB) WorldConfig {
	t.Helper()

	return WorldConfig{
		Grid: gridForTest(t, "0", 0, 0),
		Door: worldpath.Position{
			Point: pointForTest(t, 0, 0),
		},
	}
}
