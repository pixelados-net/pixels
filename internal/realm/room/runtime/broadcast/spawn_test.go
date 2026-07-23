package broadcast

import (
	"context"
	"testing"

	"github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// TestRoomSpawnSendsUnitAndStatus verifies spawn packet broadcasting.
func TestRoomSpawnSendsUnitAndStatus(t *testing.T) {
	connections := netconn.NewRegistry()
	sent := registerConnectionForTest(t, connections, "other")
	room := loadedRoomForSpawnTest(t)
	if _, err := room.Join(live.Occupant{PlayerID: 7, Username: "self", ConnectionID: "self", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("join self: %v", err)
	}
	if _, err := room.Join(live.Occupant{PlayerID: 8, Username: "other", ConnectionID: "other", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("join other: %v", err)
	}

	err := RoomSpawn(context.Background(), connections, room, 7, 7)
	if err != nil {
		t.Fatalf("spawn room unit: %v", err)
	}
	if len(*sent) != 2 || (*sent)[0].Header != 374 || (*sent)[1].Header != 1640 {
		t.Fatalf("unexpected spawn packets %#v", *sent)
	}
}

// TestRoomSpawnSkipsMissingRecords verifies empty spawn guards.
func TestRoomSpawnSkipsMissingRecords(t *testing.T) {
	room := loadedRoomForSpawnTest(t)
	err := RoomSpawn(context.Background(), netconn.NewRegistry(), room, 99, 0)
	if err != nil {
		t.Fatalf("spawn missing room unit: %v", err)
	}
}

// TestRoomUnitStatusEncodesPacket verifies focused unit status broadcasting.
func TestRoomUnitStatusEncodesPacket(t *testing.T) {
	connections := netconn.NewRegistry()
	sent := registerConnectionForTest(t, connections, "conn")
	room := loadedRoomForSpawnTest(t)
	if _, err := room.Join(live.Occupant{PlayerID: 7, Username: "demo", ConnectionID: "conn", ConnectionKind: "websocket"}); err != nil {
		t.Fatalf("join room: %v", err)
	}
	units := room.Units()
	if len(units) != 1 {
		t.Fatalf("expected one room unit %#v", units)
	}

	err := RoomUnitStatus(context.Background(), connections, room, units[0], 0)
	if err != nil {
		t.Fatalf("unit status: %v", err)
	}
	if len(*sent) != 1 || (*sent)[0].Header != 1640 {
		t.Fatalf("unexpected status packets %#v", *sent)
	}
}

// loadedRoomForSpawnTest creates a loaded room for spawn projections.
func loadedRoomForSpawnTest(t *testing.T) *live.Room {
	t.Helper()

	room, err := live.NewRoom(live.Snapshot{ID: 9, MaxUsers: 5})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	roomGrid, err := grid.Parse("0", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatalf("parse grid: %v", err)
	}
	if err := room.LoadWorld(live.WorldConfig{Grid: roomGrid, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}}); err != nil {
		t.Fatalf("load world: %v", err)
	}

	return room
}
