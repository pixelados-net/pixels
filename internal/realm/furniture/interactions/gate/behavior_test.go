package gate

import (
	"errors"
	"testing"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// TestNextRejectsOccupiedFootprint verifies every gate tile blocks a state change.
func TestNextRejectsOccupiedFootprint(t *testing.T) {
	active, gate := gateRoomForTest(t, 2, "1")
	if _, err := active.Join(roomlive.Occupant{PlayerID: 7, Username: "demo", ConnectionID: "one", ConnectionKind: netconn.Kind("websocket")}); err != nil {
		t.Fatalf("join room: %v", err)
	}
	if _, err := active.TeleportUnit(7, grid.MustPoint(2, 0), worldunit.RotationNorth, false); err != nil {
		t.Fatalf("position on second gate tile: %v", err)
	}
	next, rebuild, commit := (Behavior{}).Next(active, gate)
	if next != "1" || rebuild || commit {
		t.Fatalf("expected occupied rejection next=%q rebuild=%v commit=%v", next, rebuild, commit)
	}
}

// TestNextAlternatesFreeGate verifies both binary state directions.
func TestNextAlternatesFreeGate(t *testing.T) {
	active, item := gateRoomForTest(t, 2, "0")
	next, rebuild, commit := (Behavior{}).Next(active, item)
	if next != "1" || !rebuild || !commit {
		t.Fatalf("unexpected open transition next=%q rebuild=%v commit=%v", next, rebuild, commit)
	}
	item.ExtraData = "1"
	next, rebuild, commit = (Behavior{}).Next(active, item)
	if next != "0" || !rebuild || !commit {
		t.Fatalf("unexpected close transition next=%q rebuild=%v commit=%v", next, rebuild, commit)
	}
	item.ExtraData = ""
	next, rebuild, commit = (Behavior{}).Next(active, item)
	if next != "1" || !rebuild || !commit {
		t.Fatalf("unexpected empty transition next=%q rebuild=%v commit=%v", next, rebuild, commit)
	}
}

// TestGateStateControlsPathfinding verifies opening, closing, and stale-path cancellation.
func TestGateStateControlsPathfinding(t *testing.T) {
	active, gate := gateRoomForTest(t, 1, "1")
	if _, err := active.Join(roomlive.Occupant{PlayerID: 7, Username: "demo", ConnectionID: "one", ConnectionKind: netconn.Kind("websocket")}); err != nil {
		t.Fatalf("join room: %v", err)
	}
	if _, err := active.MoveTo(7, grid.MustPoint(2, 0)); err != nil {
		t.Fatalf("walk through open gate: %v", err)
	}
	if _, err := active.UpdateFurnitureState(gate.ID, "0", true); err != nil {
		t.Fatalf("close gate: %v", err)
	}
	movements := active.Tick()
	if len(movements) != 1 || !movements[0].Settled || movements[0].Moved {
		t.Fatalf("expected stale path cancellation, got %#v", movements)
	}
	if _, err := active.MoveTo(7, grid.MustPoint(2, 0)); !errors.Is(err, worldpath.ErrNoPath) {
		t.Fatalf("expected closed gate to block path, got %v", err)
	}
}

// gateRoomForTest creates a loaded room with one gate spanning the requested width.
func gateRoomForTest(t testing.TB, width int, state string) (*roomlive.Room, worldfurniture.Item) {
	t.Helper()
	roomGrid, err := grid.Parse("0000", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatalf("parse grid: %v", err)
	}
	active, err := roomlive.NewRoom(roomlive.Snapshot{ID: 9, OwnerPlayerID: 7, MaxUsers: 25})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	item := worldfurniture.Item{
		ID: 1, Point: grid.MustPoint(1, 0), ExtraData: state,
		Definition: worldfurniture.Definition{InteractionType: "gate", InteractionModesCount: 2, Width: width, Length: 1},
	}
	if err := active.LoadWorld(roomlive.WorldConfig{
		Grid: roomGrid, Furniture: []worldfurniture.Item{item}, Door: worldpath.Position{Point: grid.MustPoint(0, 0)},
	}); err != nil {
		t.Fatalf("load world: %v", err)
	}

	return active, item
}

// BenchmarkNextFree measures a free multi-tile gate occupancy check.
func BenchmarkNextFree(b *testing.B) {
	active, item := gateRoomForTest(b, 2, "0")
	b.ReportAllocs()
	for b.Loop() {
		_, _, _ = (Behavior{}).Next(active, item)
	}
}
