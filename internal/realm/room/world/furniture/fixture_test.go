package furniture

import (
	"testing"

	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	"github.com/niflaot/pixels/internal/realm/room/world/surface"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// TestFixturesBlocksNonWalkableFootprint verifies a blocking item occupies its whole footprint.
func TestFixturesBlocksNonWalkableFootprint(t *testing.T) {
	item := Item{
		ID:       1,
		Point:    grid.MustPoint(4, 4),
		Z:        0,
		Rotation: worldunit.RotationNorth,
		Definition: Definition{
			Width: 2, Length: 2, StackHeight: 1, AllowStack: true,
		},
	}

	fixtures, err := Fixtures(item)
	if err != nil {
		t.Fatalf("build fixtures: %v", err)
	}
	if len(fixtures) != 4 {
		t.Fatalf("expected four fixtures, got %d", len(fixtures))
	}
	for _, fixture := range fixtures {
		assertFixture(t, fixture, 0, 1, surface.StateBlocked, true, 1)
	}
}

// TestFixturesProducesSitStateForChairSlot verifies a one-slot chair.
func TestFixturesProducesSitStateForChairSlot(t *testing.T) {
	item := Item{
		ID:       2,
		Point:    grid.MustPoint(3, 3),
		Z:        0,
		Rotation: worldunit.RotationNorth,
		Definition: Definition{
			Width: 1, Length: 1, StackHeight: 1, AllowSit: true, AllowStack: false,
			Slots: []SlotDefinition{{DX: 0, DY: 0, Status: SlotStatusSit, BodyRotation: worldunit.RotationSouth}},
		},
	}

	fixtures, err := Fixtures(item)
	if err != nil {
		t.Fatalf("build fixtures: %v", err)
	}
	if len(fixtures) != 1 {
		t.Fatalf("expected one fixture, got %d", len(fixtures))
	}
	assertFixture(t, fixtures[0], 1, 1, surface.StateSit, false, 2)
}

// TestFixturesBlocksNonSlotTilesOfLayItem verifies a bed-shaped item blocks its non-slot tiles.
func TestFixturesBlocksNonSlotTilesOfLayItem(t *testing.T) {
	item := Item{
		ID:       3,
		Point:    grid.MustPoint(2, 6),
		Z:        0,
		Rotation: worldunit.RotationNorth,
		Definition: Definition{
			Width: 1, Length: 3, StackHeight: 2, AllowLay: true, AllowStack: false,
			Slots: []SlotDefinition{{DX: 0, DY: 1, Status: SlotStatusLay, BodyRotation: worldunit.RotationSouth}},
		},
	}

	fixtures, err := Fixtures(item)
	if err != nil {
		t.Fatalf("build fixtures: %v", err)
	}
	if len(fixtures) != 3 {
		t.Fatalf("expected three fixtures, got %d", len(fixtures))
	}
	assertFixture(t, fixtures[0], 0, 2, surface.StateBlocked, false, 3)
	assertFixture(t, fixtures[1], 2, 2, surface.StateLay, false, 3)
	assertFixture(t, fixtures[2], 0, 2, surface.StateBlocked, false, 3)
}

// TestFixturesOpensWalkableItem verifies a walkable, non-sit item stays open.
func TestFixturesOpensWalkableItem(t *testing.T) {
	item := Item{
		ID:       4,
		Point:    grid.MustPoint(0, 0),
		Z:        0,
		Rotation: worldunit.RotationNorth,
		Definition: Definition{
			Width: 1, Length: 1, StackHeight: 0, AllowWalk: true, AllowStack: true,
		},
	}

	fixtures, err := Fixtures(item)
	if err != nil {
		t.Fatalf("build fixtures: %v", err)
	}
	if len(fixtures) != 1 {
		t.Fatalf("expected one fixture, got %d", len(fixtures))
	}
	assertFixture(t, fixtures[0], 0, 0, surface.StateOpen, true, 4)
}

// assertFixture verifies a resolved fixture section.
func assertFixture(t *testing.T, fixture surface.Fixture, z grid.Height, top grid.Height, state surface.State, stacking bool, sourceID int64) {
	t.Helper()

	section := fixture.Section()
	if section.Z() != z || section.Top() != top || section.State() != state || section.Stacking() != stacking || section.SourceID() != sourceID {
		t.Fatalf("unexpected fixture section=%#v", section)
	}
}
