package live

import (
	"errors"
	"testing"

	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
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

// TestRoomLoadWorldProjectsFurnitureIntoFixturesAndSnapshot verifies furniture loading.
func TestRoomLoadWorldProjectsFurnitureIntoFixturesAndSnapshot(t *testing.T) {
	room, err := NewRoom(Snapshot{ID: 9, MaxUsers: 128})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	roomGrid := gridForTest(t, "000", 0, 0)
	table := worldfurniture.Item{
		ID:       11,
		Point:    pointForTest(t, 2, 0),
		Rotation: worldunit.RotationNorth,
		Definition: worldfurniture.Definition{
			Width: 1, Length: 1, StackHeight: 1,
		},
	}

	if err := room.LoadWorld(WorldConfig{
		Grid:      roomGrid,
		Furniture: []worldfurniture.Item{table},
		Door:      worldpath.Position{Point: pointForTest(t, 0, 0)},
		Body:      worldunit.RotationSouth,
		Head:      worldunit.RotationSouth,
	}); err != nil {
		t.Fatalf("load world: %v", err)
	}

	items := room.FurnitureItems()
	if len(items) != 1 || items[0].ID != 11 {
		t.Fatalf("unexpected furniture snapshot %#v", items)
	}

	_, err = room.ResolveFurniturePlacement(99, []grid.Point{pointForTest(t, 2, 0)})
	if !errors.Is(err, ErrCannotStack) {
		t.Fatalf("expected blocked furniture surface, got %v", err)
	}

	if _, err := room.Join(occupantForTest(7)); err != nil {
		t.Fatalf("join room: %v", err)
	}
	if _, err := room.MoveTo(7, pointForTest(t, 2, 0)); !errors.Is(err, worldpath.ErrNoPath) {
		t.Fatalf("expected blocked tile to be unreachable, got %v", err)
	}
}

// TestRoomSurfaceHeightsReturnsZeroWithoutWorld verifies the unloaded-world default.
func TestRoomSurfaceHeightsReturnsZeroWithoutWorld(t *testing.T) {
	room, err := NewRoom(Snapshot{ID: 9, MaxUsers: 2})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}

	width, height, tiles := room.SurfaceHeights()
	if width != 0 || height != 0 || tiles != nil {
		t.Fatalf("expected zero surface heights, got width=%d height=%d tiles=%#v", width, height, tiles)
	}
}

// TestRoomSurfaceHeightsResolvesValidAndInvalidTiles verifies per-tile height resolution.
func TestRoomSurfaceHeightsResolvesValidAndInvalidTiles(t *testing.T) {
	room := worldRoomForTest(t, "0x", 0, 0)

	width, height, tiles := room.SurfaceHeights()
	if width != 2 || height != 1 || len(tiles) != 2 {
		t.Fatalf("unexpected surface dimensions width=%d height=%d tiles=%#v", width, height, tiles)
	}
	if !tiles[0].Valid || tiles[0].Height != 0 || tiles[0].StackingBlocked {
		t.Fatalf("unexpected valid tile %#v", tiles[0])
	}
	if tiles[1].Valid {
		t.Fatalf("expected invalid tile, got %#v", tiles[1])
	}
}

// TestRoomFurnitureItemsReturnsNilWithoutWorld verifies the snapshot default.
func TestRoomFurnitureItemsReturnsNilWithoutWorld(t *testing.T) {
	room, err := NewRoom(Snapshot{ID: 9, MaxUsers: 2})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}

	if items := room.FurnitureItems(); items != nil {
		t.Fatalf("expected nil furniture snapshot, got %#v", items)
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
