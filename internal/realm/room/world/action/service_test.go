package action

import (
	"context"
	"testing"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	"github.com/niflaot/pixels/pkg/bus"
)

// TestServiceCoordinatesAvatarState verifies dance, expression, idle, and posture invariants.
func TestServiceCoordinatesAvatarState(t *testing.T) {
	active := actionRoom(t)
	service := New(nil, bus.New())
	active.SetUnitStatus(7, worldunit.StatusSit, "0.5")
	if err := service.Dance(context.Background(), active, 7, 3); err != nil {
		t.Fatal(err)
	}
	unit, _ := active.Unit(7)
	if hasStatus(unit, worldunit.StatusSit) || !hasStatus(unit, worldunit.StatusDance) {
		t.Fatalf("unexpected dancing unit %#v", unit)
	}
	if err := service.Express(context.Background(), active, 7, 2); err != nil {
		t.Fatal(err)
	}
	if err := service.SetIdle(context.Background(), active, 7, true); err != nil {
		t.Fatal(err)
	}
	unit, _ = active.Unit(7)
	if !unit.Idle || hasStatus(unit, worldunit.StatusDance) {
		t.Fatalf("unexpected idle unit %#v", unit)
	}
	if err := service.Dance(context.Background(), active, 7, 4); err != nil {
		t.Fatal(err)
	}
	unit, _ = active.Unit(7)
	if unit.Idle || !hasStatus(unit, worldunit.StatusDance) {
		t.Fatalf("unexpected resumed unit %#v", unit)
	}
	if err := service.Posture(context.Background(), active, 7, true); err != nil {
		t.Fatal(err)
	}
	unit, _ = active.Unit(7)
	if hasStatus(unit, worldunit.StatusDance) || !hasStatus(unit, worldunit.StatusSit) {
		t.Fatalf("unexpected sitting unit %#v", unit)
	}
}

// TestServiceMissingUnitReturnsStableErrors verifies stale room commands are harmless.
func TestServiceMissingUnitReturnsStableErrors(t *testing.T) {
	active := actionRoom(t)
	service := New(nil, nil)
	if err := service.Dance(context.Background(), active, 99, 1); err != roomlive.ErrUnitNotFound {
		t.Fatalf("expected missing dance unit, got %v", err)
	}
	if err := service.Express(context.Background(), active, 99, 1); err != roomlive.ErrUnitNotFound {
		t.Fatalf("expected missing expression unit, got %v", err)
	}
	if err := service.SetIdle(context.Background(), active, 99, true); err != roomlive.ErrUnitNotFound {
		t.Fatalf("expected missing idle unit, got %v", err)
	}
}

// actionRoom creates one loaded room with a player unit.
func actionRoom(t testing.TB) *roomlive.Room {
	t.Helper()
	roomGrid, err := grid.Parse("00", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	active, err := roomlive.NewRoom(roomlive.Snapshot{ID: 9, MaxUsers: 5})
	if err != nil {
		t.Fatal(err)
	}
	if err = active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}, Body: worldunit.RotationSouth, Head: worldunit.RotationSouth}); err != nil {
		t.Fatal(err)
	}
	if _, err = active.Join(roomlive.Occupant{PlayerID: 7, Username: "demo", ConnectionID: "action", ConnectionKind: "websocket"}); err != nil {
		t.Fatal(err)
	}
	return active
}
