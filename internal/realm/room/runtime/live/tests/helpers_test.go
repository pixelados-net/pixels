package live_test

import (
	"testing"

	. "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	"github.com/niflaot/pixels/internal/realm/room/world/surface"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// occupantForTest creates a valid room occupant.
func occupantForTest(playerID int64) Occupant {
	return Occupant{
		PlayerID: playerID, Username: "player", ConnectionID: netconn.ID("connection"),
		ConnectionKind: netconn.Kind("websocket"),
	}
}

// worldRoomForTest creates a room with loaded world behavior.
func worldRoomForTest(t testing.TB, heightmap string, doorX int, doorY int) *Room {
	t.Helper()
	room, err := NewRoom(Snapshot{ID: 9, MaxUsers: 128})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	if err := room.LoadWorld(WorldConfig{
		Grid: gridForTest(t, heightmap, doorX, doorY),
		Door: worldpath.Position{Point: pointForTest(t, doorX, doorY)},
		Body: worldunit.RotationSouth, Head: worldunit.RotationSouth,
	}); err != nil {
		t.Fatalf("load world: %v", err)
	}

	return room
}

// gridForTest parses a room grid.
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

// fixtureForLiveTest creates a surface fixture for live behavior tests.
func fixtureForLiveTest(t *testing.T, params surface.FixtureParams) surface.Fixture {
	t.Helper()
	fixture, err := surface.NewFixture(params)
	if err != nil {
		t.Fatalf("create fixture: %v", err)
	}

	return fixture
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

// hasStatusValue reports whether a status key holds an exact value.
func hasStatusValue(statuses []worldunit.Status, key string, value string) bool {
	for _, status := range statuses {
		if status.Key == key && status.Value == value {
			return true
		}
	}

	return false
}
