package live

import (
	"errors"
	"testing"

	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	"github.com/niflaot/pixels/internal/realm/room/world/surface"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// TestRoomMoveToAndTickAdvancesUnit verifies runtime movement ticks.
func TestRoomMoveToAndTickAdvancesUnit(t *testing.T) {
	room := worldRoomForTest(t, "000", 0, 0)
	if _, err := room.Join(occupantForTest(7)); err != nil {
		t.Fatalf("join room: %v", err)
	}
	path, err := room.MoveTo(7, pointForTest(t, 2, 0))
	if err != nil {
		t.Fatalf("move unit: %v", err)
	}
	if path.Len() != 2 {
		t.Fatalf("unexpected path length %d", path.Len())
	}

	first := room.Tick()
	if len(first) != 1 || first[0].Unit.Position.Point != pointForTest(t, 1, 0) {
		t.Fatalf("unexpected first tick %#v", first)
	}
	if !hasStatus(first[0].Unit.Statuses, worldunit.StatusMove) {
		t.Fatalf("expected move status %#v", first[0].Unit.Statuses)
	}

	second := room.Tick()
	if len(second) != 1 || second[0].Unit.Position.Point != pointForTest(t, 2, 0) {
		t.Fatalf("unexpected second tick %#v", second)
	}
	if second[0].Unit.Moving || !second[0].Moved || second[0].Settled {
		t.Fatalf("expected movement completed %#v", second[0].Unit)
	}
	if !hasStatus(second[0].Unit.Statuses, worldunit.StatusMove) {
		t.Fatalf("expected final move status %#v", second[0].Unit.Statuses)
	}

	third := room.Tick()
	if len(third) != 1 || third[0].Moved || !third[0].Settled {
		t.Fatalf("unexpected settle tick %#v", third)
	}
	if hasStatus(third[0].Unit.Statuses, worldunit.StatusMove) {
		t.Fatalf("expected clean settled status %#v", third[0].Unit.Statuses)
	}
}

// TestRoomFaceToClearsMovementAndRotates verifies facing an occupied target.
func TestRoomFaceToClearsMovementAndRotates(t *testing.T) {
	room := worldRoomForTest(t, "00", 0, 0)
	if _, err := room.Join(occupantForTest(7)); err != nil {
		t.Fatalf("join room: %v", err)
	}
	if _, err := room.MoveTo(7, pointForTest(t, 1, 0)); err != nil {
		t.Fatalf("move unit: %v", err)
	}

	unit, err := room.FaceTo(7, pointForTest(t, 1, 0))
	if err != nil {
		t.Fatalf("face unit: %v", err)
	}
	if unit.Moving || unit.BodyRotation != worldunit.RotationEast || unit.HeadRotation != worldunit.RotationEast {
		t.Fatalf("unexpected faced unit %#v", unit)
	}
}

// TestRoomMoveToAvoidsOccupiedUnit verifies occupancy-aware paths.
func TestRoomMoveToAvoidsOccupiedUnit(t *testing.T) {
	room := worldRoomForTest(t, "000\r000\r000", 0, 1)
	if _, err := room.Join(occupantForTest(7)); err != nil {
		t.Fatalf("join first room unit: %v", err)
	}
	if _, err := room.Join(occupantForTest(8)); err != nil {
		t.Fatalf("join second room unit: %v", err)
	}
	if _, err := room.MoveTo(8, pointForTest(t, 1, 1)); err != nil {
		t.Fatalf("move blocker: %v", err)
	}
	if movements := room.Tick(); len(movements) != 1 {
		t.Fatalf("expected blocker movement %#v", movements)
	}

	path, err := room.MoveTo(7, pointForTest(t, 2, 1))
	if err != nil {
		t.Fatalf("move around blocker: %v", err)
	}
	for _, step := range path.Steps() {
		if step.Position.Point == pointForTest(t, 1, 1) {
			t.Fatalf("path stepped into occupied tile %#v", path.Steps())
		}
	}
}

// TestRoomMoveToAvoidsReservedGoal verifies pending movement targets are blocked.
func TestRoomMoveToAvoidsReservedGoal(t *testing.T) {
	room := worldRoomForTest(t, "000", 0, 0)
	if _, err := room.Join(occupantForTest(7)); err != nil {
		t.Fatalf("join first room unit: %v", err)
	}
	if _, err := room.Join(occupantForTest(8)); err != nil {
		t.Fatalf("join second room unit: %v", err)
	}
	if _, err := room.MoveTo(8, pointForTest(t, 1, 0)); err != nil {
		t.Fatalf("move second room unit: %v", err)
	}
	if movements := room.Tick(); len(movements) != 1 {
		t.Fatalf("expected second movement %#v", movements)
	}
	if _, err := room.MoveTo(8, pointForTest(t, 2, 0)); err != nil {
		t.Fatalf("reserve second goal: %v", err)
	}

	_, err := room.MoveTo(7, pointForTest(t, 2, 0))
	if !errors.Is(err, worldpath.ErrNoPath) {
		t.Fatalf("expected reserved goal to block path, got %v", err)
	}
}

// TestRoomTickAppliesSitStatusOnSettle verifies seat resolution after settling.
func TestRoomTickAppliesSitStatusOnSettle(t *testing.T) {
	seat := pointForTest(t, 1, 0)
	room := worldRoomForTest(t, "00", 0, 0)
	if _, err := room.Join(occupantForTest(7)); err != nil {
		t.Fatalf("join room: %v", err)
	}
	chair := worldfurniture.Item{
		ID: 5, Point: seat, Rotation: worldunit.RotationNorth,
		Definition: worldfurniture.Definition{
			Width: 1, Length: 1, StackHeight: 1, AllowStack: true, AllowSit: true,
			Slots: []worldfurniture.SlotDefinition{{Status: worldfurniture.SlotStatusSit}},
		},
	}
	if _, err := room.ReloadFurniture(5, &chair); err != nil {
		t.Fatalf("place chair: %v", err)
	}
	if _, err := room.MoveTo(7, seat); err != nil {
		t.Fatalf("move to seat: %v", err)
	}

	first := room.Tick()
	if len(first) != 1 || !first[0].Moved {
		t.Fatalf("unexpected first tick %#v", first)
	}

	second := room.Tick()
	if len(second) != 1 || !second[0].Settled {
		t.Fatalf("unexpected settle tick %#v", second)
	}
	if !hasStatusValue(second[0].Unit.Statuses, worldunit.StatusSit, "1") {
		t.Fatalf("expected sit status at height 1, got %#v", second[0].Unit.Statuses)
	}
}

// TestRoomReloadFixturesPreservesUnits verifies fixture updates do not reset units.
func TestRoomReloadFixturesPreservesUnits(t *testing.T) {
	room := worldRoomForTest(t, "000", 0, 0)
	if _, err := room.Join(occupantForTest(7)); err != nil {
		t.Fatalf("join room: %v", err)
	}
	if _, err := room.MoveTo(7, pointForTest(t, 2, 0)); err != nil {
		t.Fatalf("move unit: %v", err)
	}
	room.Tick()
	room.Tick()

	before := room.Units()
	if len(before) != 1 {
		t.Fatalf("expected one unit before reload, got %#v", before)
	}
	position := before[0].Position

	blocking := fixtureForLiveTest(t, surface.FixtureParams{Point: pointForTest(t, 0, 0), Z: 5, Top: 5, State: surface.StateBlocked, SourceID: 11})
	if err := room.ReloadFixtures(11, []surface.Fixture{blocking}); err != nil {
		t.Fatalf("reload fixtures: %v", err)
	}

	after := room.Units()
	if len(after) != 1 || after[0].Position != position {
		t.Fatalf("expected unit position preserved, got %#v", after)
	}
}

// TestRoomTickCancelsPathWhenFixturesChangeMidWalk verifies stale path invalidation.
func TestRoomTickCancelsPathWhenFixturesChangeMidWalk(t *testing.T) {
	room := worldRoomForTest(t, "0000", 0, 0)
	if _, err := room.Join(occupantForTest(7)); err != nil {
		t.Fatalf("join room: %v", err)
	}
	if _, err := room.MoveTo(7, pointForTest(t, 3, 0)); err != nil {
		t.Fatalf("move unit: %v", err)
	}

	blocking := fixtureForLiveTest(t, surface.FixtureParams{Point: pointForTest(t, 2, 0), Z: 5, Top: 5, State: surface.StateBlocked, SourceID: 3})
	if err := room.ReloadFixtures(3, []surface.Fixture{blocking}); err != nil {
		t.Fatalf("reload fixtures: %v", err)
	}

	movements := room.Tick()
	if len(movements) != 1 || movements[0].Moved || !movements[0].Settled || movements[0].Unit.Moving {
		t.Fatalf("expected neutral stop movement, got %#v", movements)
	}
	if hasStatus(movements[0].Unit.Statuses, worldunit.StatusMove) {
		t.Fatalf("expected move status cleared, got %#v", movements[0].Unit.Statuses)
	}

	units := room.Units()
	if len(units) != 1 || units[0].Moving {
		t.Fatalf("expected cleared movement, got %#v", units)
	}
}

// TestRoomReloadFurnitureTracksSnapshotAndFixtures verifies furniture reload updates both resolver and snapshot.
func TestRoomReloadFurnitureTracksSnapshotAndFixtures(t *testing.T) {
	room := worldRoomForTest(t, "000", 0, 0)
	item := worldfurniture.Item{
		ID:       5,
		Point:    pointForTest(t, 2, 0),
		Rotation: worldunit.RotationNorth,
		Definition: worldfurniture.Definition{
			Width: 1, Length: 1, StackHeight: 1,
		},
	}

	if _, err := room.ReloadFurniture(5, &item); err != nil {
		t.Fatalf("reload furniture: %v", err)
	}
	if items := room.FurnitureItems(); len(items) != 1 || items[0].ID != 5 {
		t.Fatalf("unexpected furniture snapshot %#v", items)
	}

	_, err := room.ResolveFurniturePlacement(99, []grid.Point{pointForTest(t, 2, 0)})
	if !errors.Is(err, ErrCannotStack) {
		t.Fatalf("expected blocked surface from reloaded furniture, got %v", err)
	}

	if _, err := room.ReloadFurniture(5, nil); err != nil {
		t.Fatalf("remove furniture: %v", err)
	}
	if items := room.FurnitureItems(); len(items) != 0 {
		t.Fatalf("expected furniture removed, got %#v", items)
	}
}
