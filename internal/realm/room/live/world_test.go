package live

import (
	"errors"
	"strings"
	"testing"

	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// TestRoomLoadWorldCreatesUnitsForJoin verifies world unit bootstrap.
func TestRoomLoadWorldCreatesUnitsForJoin(t *testing.T) {
	room := worldRoomForTest(t, "0", 0, 0)
	if _, err := room.Join(occupantForTest(7)); err != nil {
		t.Fatalf("join room: %v", err)
	}

	units := room.Units()
	if len(units) != 1 || units[0].PlayerID != 7 || units[0].UnitID != 1 {
		t.Fatalf("unexpected units %#v", units)
	}
	if units[0].Position.Point != pointForTest(t, 0, 0) || units[0].Position.Z != 0 {
		t.Fatalf("unexpected position %#v", units[0].Position)
	}
}

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
	if second[0].Unit.Moving {
		t.Fatalf("expected movement completed %#v", second[0].Unit)
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

// TestRoomMoveToRejectsMissingWorldOrUnit verifies movement validation.
func TestRoomMoveToRejectsMissingWorldOrUnit(t *testing.T) {
	room, err := NewRoom(Snapshot{ID: 9, MaxUsers: 2})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}

	_, err = room.MoveTo(7, pointForTest(t, 0, 0))
	if !errors.Is(err, ErrWorldNotLoaded) {
		t.Fatalf("expected world not loaded, got %v", err)
	}

	room = worldRoomForTest(t, "0", 0, 0)
	_, err = room.MoveTo(7, pointForTest(t, 0, 0))
	if !errors.Is(err, ErrUnitNotFound) {
		t.Fatalf("expected unit not found, got %v", err)
	}
}

// TestRoomUnloadWorldClearsUnitSnapshots verifies world unloading.
func TestRoomUnloadWorldClearsUnitSnapshots(t *testing.T) {
	room := worldRoomForTest(t, "0", 0, 0)
	if _, err := room.Join(occupantForTest(7)); err != nil {
		t.Fatalf("join room: %v", err)
	}

	room.UnloadWorld()
	if room.WorldLoaded() || len(room.Units()) != 0 {
		t.Fatalf("expected unloaded world")
	}
}

// TestRoomLoadWorldRejectsInvalidDoor verifies world door validation.
func TestRoomLoadWorldRejectsInvalidDoor(t *testing.T) {
	room, err := NewRoom(Snapshot{ID: 9, MaxUsers: 2})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	roomGrid := gridForTest(t, "0", 0, 0)

	err = room.LoadWorld(WorldConfig{
		Grid: roomGrid,
		Door: worldpath.Position{Point: pointForTest(t, 0, 0), Z: 2},
	})
	if !errors.Is(err, ErrInvalidWorld) {
		t.Fatalf("expected invalid world, got %v", err)
	}
}

// worldRoomForTest creates a room with loaded world behavior.
func worldRoomForTest(t testing.TB, heightmap string, doorX int, doorY int) *Room {
	t.Helper()

	room, err := NewRoom(Snapshot{ID: 9, MaxUsers: 128})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	roomGrid := gridForTest(t, heightmap, doorX, doorY)
	if err := room.LoadWorld(WorldConfig{
		Grid: roomGrid,
		Door: worldpath.Position{Point: pointForTest(t, doorX, doorY)},
		Body: worldunit.RotationSouth,
		Head: worldunit.RotationSouth,
	}); err != nil {
		t.Fatalf("load world: %v", err)
	}

	return room
}

// gridForTest creates a parsed test grid.
func gridForTest(t testing.TB, heightmap string, doorX int, doorY int) grid.Grid {
	t.Helper()

	roomGrid, err := grid.Parse(heightmap, grid.WithDoor(doorX, doorY))
	if err != nil {
		t.Fatalf("parse grid: %v", err)
	}

	return roomGrid
}

// pointForTest creates a test grid point.
func pointForTest(t testing.TB, x int, y int) grid.Point {
	t.Helper()

	point, ok := grid.NewPoint(x, y)
	if !ok {
		t.Fatalf("invalid point %d,%d", x, y)
	}

	return point
}

// hasStatus reports whether a status key exists.
func hasStatus(statuses []worldunit.Status, key string) bool {
	for _, status := range statuses {
		if status.Key == key {
			return true
		}
	}

	return false
}

// BenchmarkRoomMoveTo measures runtime path assignment cost.
func BenchmarkRoomMoveTo(b *testing.B) {
	room := worldRoomForTest(b, heightmapForBenchmark(24), 0, 0)
	if _, err := room.Join(occupantForTest(7)); err != nil {
		b.Fatalf("join room: %v", err)
	}
	goal := pointForTest(b, 23, 23)

	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		if _, err := room.MoveTo(7, goal); err != nil {
			b.Fatalf("move unit: %v", err)
		}
	}
}

// BenchmarkRoomTickManyUnits measures one tick over many moving units.
func BenchmarkRoomTickManyUnits(b *testing.B) {
	room := worldRoomForTest(b, heightmapForBenchmark(16), 0, 0)
	for playerID := int64(1); playerID <= 64; playerID++ {
		if _, err := room.Join(occupantForTest(playerID)); err != nil {
			b.Fatalf("join room: %v", err)
		}
	}

	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		room.mutex.Lock()
		for playerID := int64(1); playerID <= 64; playerID++ {
			room.world.units[playerID].SetPath(worldpath.NewPath([]worldpath.Step{{
				Position: worldpath.Position{Point: pointForTest(b, int((playerID+int64(index))%16), int(playerID/16))},
			}}))
		}
		room.mutex.Unlock()
		_ = room.Tick()
	}
}

// heightmapForBenchmark creates a square flat heightmap.
func heightmapForBenchmark(size int) string {
	row := strings.Repeat("0", size)
	rows := make([]string, size)
	for index := range rows {
		rows[index] = row
	}

	return strings.Join(rows, "\r")
}
