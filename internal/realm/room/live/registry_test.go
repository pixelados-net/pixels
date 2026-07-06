package live

import (
	"context"
	"errors"
	"testing"
	"time"

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

// occupantForTest creates an active room occupant.
func occupantForTest(playerID int64) Occupant {
	return Occupant{PlayerID: playerID, Username: "demo", ConnectionID: netconn.ID("conn"), ConnectionKind: netconn.Kind("websocket")}
}
