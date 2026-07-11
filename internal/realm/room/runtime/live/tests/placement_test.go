package live_test

import (
	"errors"
	"testing"

	. "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	"github.com/niflaot/pixels/internal/realm/room/world/surface"
)

// TestResolveFurniturePlacementOnEmptyFloorReturnsBaseHeight verifies the flat-room baseline.
func TestResolveFurniturePlacementOnEmptyFloorReturnsBaseHeight(t *testing.T) {
	room := worldRoomForTest(t, "000", 0, 0)

	height, err := room.ResolveFurniturePlacement(1, []grid.Point{pointForTest(t, 2, 0)})
	if err != nil {
		t.Fatalf("resolve placement: %v", err)
	}
	if height != 0 {
		t.Fatalf("expected base height 0, got %d", height)
	}
}

// TestResolveFurniturePlacementRejectsInvalidTile verifies out-of-grid footprints are rejected.
func TestResolveFurniturePlacementRejectsInvalidTile(t *testing.T) {
	room := worldRoomForTest(t, "000", 0, 0)

	_, err := room.ResolveFurniturePlacement(1, []grid.Point{pointForTest(t, 9, 9)})
	if !errors.Is(err, ErrInvalidPlacement) {
		t.Fatalf("expected invalid placement, got %v", err)
	}
}

// TestResolveFurniturePlacementRejectsOccupiedTile verifies units block placement.
func TestResolveFurniturePlacementRejectsOccupiedTile(t *testing.T) {
	room := worldRoomForTest(t, "000", 0, 0)
	if _, err := room.Join(occupantForTest(7)); err != nil {
		t.Fatalf("join room: %v", err)
	}

	_, err := room.ResolveFurniturePlacement(1, []grid.Point{pointForTest(t, 0, 0)})
	if !errors.Is(err, ErrTileOccupied) {
		t.Fatalf("expected tile occupied, got %v", err)
	}
}

// TestResolveFurniturePlacementRejectsNonStackingTop verifies stacking rules are enforced.
func TestResolveFurniturePlacementRejectsNonStackingTop(t *testing.T) {
	room := worldRoomForTest(t, "000", 0, 0)
	blocking := fixtureForLiveTest(t, surface.FixtureParams{Point: pointForTest(t, 2, 0), Z: 0, Top: 1, State: surface.StateBlocked, Stacking: false, SourceID: 9})
	if err := room.ReloadFixtures(9, []surface.Fixture{blocking}); err != nil {
		t.Fatalf("reload fixtures: %v", err)
	}

	_, err := room.ResolveFurniturePlacement(1, []grid.Point{pointForTest(t, 2, 0)})
	if !errors.Is(err, ErrCannotStack) {
		t.Fatalf("expected cannot stack, got %v", err)
	}
}

// TestResolveFurniturePlacementExcludesOwnFixtures verifies moving/rotating in place ignores itself.
func TestResolveFurniturePlacementExcludesOwnFixtures(t *testing.T) {
	room := worldRoomForTest(t, "000", 0, 0)
	blocking := fixtureForLiveTest(t, surface.FixtureParams{Point: pointForTest(t, 2, 0), Z: 0, Top: 1, State: surface.StateBlocked, Stacking: false, SourceID: 9})
	if err := room.ReloadFixtures(9, []surface.Fixture{blocking}); err != nil {
		t.Fatalf("reload fixtures: %v", err)
	}

	height, err := room.ResolveFurniturePlacement(9, []grid.Point{pointForTest(t, 2, 0)})
	if err != nil {
		t.Fatalf("expected placement excluding own fixture to succeed, got %v", err)
	}
	if height != 0 {
		t.Fatalf("expected base height 0 once own fixture is excluded, got %d", height)
	}
}

// TestResolveFurniturePlacementRejectsWithoutWorld verifies the unloaded-world guard.
func TestResolveFurniturePlacementRejectsWithoutWorld(t *testing.T) {
	room, err := NewRoom(Snapshot{ID: 9, MaxUsers: 2})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}

	_, err = room.ResolveFurniturePlacement(1, nil)
	if !errors.Is(err, ErrWorldNotLoaded) {
		t.Fatalf("expected world not loaded, got %v", err)
	}
}
