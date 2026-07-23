package floorplan

import (
	"testing"

	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
)

// TestOccupiedTilesDeduplicatesAndSortsFootprints verifies editor tile projection.
func TestOccupiedTilesDeduplicatesAndSortsFootprints(t *testing.T) {
	items := []worldfurniture.Item{
		{ID: 2, Point: grid.MustPoint(1, 0), Definition: worldfurniture.Definition{Width: 2, Length: 1}},
		{ID: 1, Point: grid.MustPoint(0, 0), Definition: worldfurniture.Definition{Width: 2, Length: 1}},
	}
	points := OccupiedTiles(items)
	if len(points) != 3 || points[0] != grid.MustPoint(0, 0) || points[2] != grid.MustPoint(2, 0) {
		t.Fatalf("unexpected occupied tiles %#v", points)
	}
}

// TestBlockedItemsDetectsHeightChangesAndRemovedTiles verifies complete support checks.
func TestBlockedItemsDetectsHeightChangesAndRemovedTiles(t *testing.T) {
	previous, err := grid.Parse("00", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatalf("parse previous: %v", err)
	}
	next, err := grid.Parse("1x", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatalf("parse next: %v", err)
	}
	items := []worldfurniture.Item{{ID: 7, Point: grid.MustPoint(0, 0), Definition: worldfurniture.Definition{Width: 2, Length: 1}}}
	blocked := BlockedItems(previous, next, items)
	if len(blocked) != 1 || blocked[0].Item.ID != 7 {
		t.Fatalf("unexpected blocked furniture %#v", blocked)
	}
}

// BenchmarkBlockedItems measures realistic furniture support validation.
func BenchmarkBlockedItems(b *testing.B) {
	previous, _ := grid.Parse("00000000\r00000000\r00000000\r00000000", grid.WithDoor(0, 0))
	next, _ := grid.Parse("11111111\r11111111\r11111111\r11111111", grid.WithDoor(0, 0))
	items := make([]worldfurniture.Item, 0, 32)
	for y := 0; y < 4; y++ {
		for x := 0; x < 8; x++ {
			items = append(items, worldfurniture.Item{ID: int64(y*8 + x + 1), Point: grid.MustPoint(x, y), Definition: worldfurniture.Definition{Width: 1, Length: 1}})
		}
	}
	for b.Loop() {
		_ = BlockedItems(previous, next, items)
	}
}
