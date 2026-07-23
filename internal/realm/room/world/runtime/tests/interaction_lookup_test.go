package tests

import (
	"testing"

	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldruntime "github.com/niflaot/pixels/internal/realm/room/world/runtime"
)

// TestNearestFurnitureByInteractionSelectsDistanceThenID verifies stable indexed selection.
func TestNearestFurnitureByInteractionSelectsDistanceThenID(t *testing.T) {
	world := interactionWorld(t)
	item, found := world.NearestFurnitureByInteraction("pet_food", grid.MustPoint(4, 0), 5)
	if !found || item.ID != 10 {
		t.Fatalf("unexpected nearest item %#v found=%v", item, found)
	}
	if _, found = world.NearestFurnitureByInteraction("pet_food", grid.MustPoint(4, 0), 0); found {
		t.Fatal("expected radius to reject every item")
	}
}

// TestNearestFurnitureByInteractionAllocatesNothing verifies the pet-need hot lookup.
func TestNearestFurnitureByInteractionAllocatesNothing(t *testing.T) {
	world := interactionWorld(t)
	allocations := testing.AllocsPerRun(1000, func() {
		_, _ = world.NearestFurnitureByInteraction("pet_food", grid.MustPoint(4, 0), 5)
	})
	if allocations != 0 {
		t.Fatalf("expected zero allocations, got %.2f", allocations)
	}
}

// BenchmarkNearestFurnitureByInteraction measures indexed autonomous-need lookup.
func BenchmarkNearestFurnitureByInteraction(b *testing.B) {
	benchmarkNearestFurnitureByInteraction(b)
}

// BenchmarkPetSpatialNeedLookup measures the PETS plan's indexed need lookup budget.
func BenchmarkPetSpatialNeedLookup(b *testing.B) {
	benchmarkNearestFurnitureByInteraction(b)
}

// benchmarkNearestFurnitureByInteraction measures one warmed interaction lookup.
func benchmarkNearestFurnitureByInteraction(b *testing.B) {
	world := interactionWorld(b)
	origin := grid.MustPoint(4, 0)
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		_, _ = world.NearestFurnitureByInteraction("pet_food", origin, 5)
	}
}

// interactionWorld creates one flat world with stable need fixtures.
func interactionWorld(t testing.TB) *worldruntime.World {
	t.Helper()
	roomGrid, err := grid.Parse("000000", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	definition := worldfurniture.Definition{InteractionType: "pet_food", Width: 1, Length: 1, StackHeight: grid.HeightFromUnits(0.1), AllowWalk: true, AllowStack: true}
	world, err := worldruntime.New(worldruntime.Config{
		Grid: roomGrid, Door: worldpath.Position{Point: grid.MustPoint(0, 0)},
		Furniture: []worldfurniture.Item{
			{ID: 20, Definition: definition, Point: grid.MustPoint(5, 0)},
			{ID: 10, Definition: definition, Point: grid.MustPoint(3, 0)},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return world
}
