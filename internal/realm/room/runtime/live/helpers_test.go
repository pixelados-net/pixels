package live

import (
	"testing"

	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	"github.com/niflaot/pixels/internal/realm/room/world/surface"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// hasStatusValue reports whether a status key holds an exact value.
func hasStatusValue(statuses []worldunit.Status, key string, value string) bool {
	for _, status := range statuses {
		if status.Key == key && status.Value == value {
			return true
		}
	}

	return false
}

// fixtureForLiveTest creates a surface fixture for live package tests.
func fixtureForLiveTest(t *testing.T, params surface.FixtureParams) surface.Fixture {
	t.Helper()
	fixture, err := surface.NewFixture(params)
	if err != nil {
		t.Fatalf("create fixture: %v", err)
	}

	return fixture
}

// worldRoomWithFixturesForTest creates a room with loaded world behavior and fixtures.
func worldRoomWithFixturesForTest(t testing.TB, heightmap string, doorX int, doorY int, fixtures []surface.Fixture) *Room {
	t.Helper()
	room, err := NewRoom(Snapshot{ID: 9, MaxUsers: 128})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}
	if err := room.LoadWorld(WorldConfig{
		Grid: gridForTest(t, heightmap, doorX, doorY), Fixtures: fixtures,
		Door: worldpath.Position{Point: pointForTest(t, doorX, doorY)},
		Body: worldunit.RotationSouth, Head: worldunit.RotationSouth,
	}); err != nil {
		t.Fatalf("load world: %v", err)
	}

	return room
}
