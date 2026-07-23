package tests

import (
	"testing"

	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldruntime "github.com/niflaot/pixels/internal/realm/room/world/runtime"
)

// TestUnitByIDTracksLifecycle verifies the reverse unit index cannot become stale.
func TestUnitByIDTracksLifecycle(t *testing.T) {
	world := lookupWorld(t, 64)
	world.AddUnit(10)
	unit, found := world.Unit(10)
	if !found {
		t.Fatal("expected player unit")
	}
	resolved, found := world.UnitByID(unit.UnitID)
	if !found || resolved.PlayerID != 10 {
		t.Fatalf("unexpected reverse lookup %#v found=%v", resolved, found)
	}
	world.RemoveUnit(10)
	if _, found = world.UnitByID(unit.UnitID); found {
		t.Fatal("removed unit remained indexed")
	}
}

// TestUnitByIDAllocatesNothing verifies the room-entity lookup hot path.
func TestUnitByIDAllocatesNothing(t *testing.T) {
	world := lookupWorld(t, 128)
	for playerID := int64(1); playerID <= 100; playerID++ {
		world.AddUnit(playerID)
	}
	allocations := testing.AllocsPerRun(1000, func() {
		_, _ = world.UnitByID(75)
	})
	if allocations != 0 {
		t.Fatalf("expected zero allocations, got %.2f", allocations)
	}
}

// BenchmarkUnitByID measures room-local target resolution for hand-item transfer.
func BenchmarkUnitByID(b *testing.B) {
	world := lookupWorld(b, 256)
	for playerID := int64(1); playerID <= 200; playerID++ {
		world.AddUnit(playerID)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		_, _ = world.UnitByID(150)
	}
}

// lookupWorld creates a flat world for unit-index tests.
func lookupWorld(t testing.TB, width int) *worldruntime.World {
	t.Helper()
	heightmap := make([]byte, width)
	for index := range heightmap {
		heightmap[index] = '0'
	}
	roomGrid, err := grid.Parse(string(heightmap), grid.WithDoor(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	world, err := worldruntime.New(worldruntime.Config{Grid: roomGrid, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}})
	if err != nil {
		t.Fatal(err)
	}
	return world
}
