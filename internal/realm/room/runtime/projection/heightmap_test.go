package projection

import (
	"testing"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
)

// TestHeightMapTilesEncodesKnownCases verifies height, stacking, and invalid tile encoding.
func TestHeightMapTilesEncodesKnownCases(t *testing.T) {
	tiles := []roomlive.TileHeight{
		{Valid: true, Height: 0, StackingBlocked: false},
		{Valid: true, Height: 1, StackingBlocked: false},
		{Valid: true, Height: 1, StackingBlocked: true},
		{},
	}

	values := HeightMapTiles(tiles)
	want := []int16{0, 256, int16(256 | 0x4000), -1}
	if len(values) != len(want) {
		t.Fatalf("unexpected value count %d", len(values))
	}
	for index, expected := range want {
		if values[index] != expected {
			t.Fatalf("tile %d: got %d, want %d", index, values[index], expected)
		}
	}
}

// TestHeightMapUpdateTilesSelectsRequestedPointsAndDedupes verifies only the requested points are
// returned, duplicates collapse to one record, and out-of-range points are skipped.
func TestHeightMapUpdateTilesSelectsRequestedPointsAndDedupes(t *testing.T) {
	width := uint16(2)
	tiles := []roomlive.TileHeight{
		{Valid: true, Height: 0},
		{Valid: true, Height: 1},
		{Valid: true, Height: 2, StackingBlocked: true},
		{},
	}
	pointA := grid.Point{X: 0, Y: 0}
	pointB := grid.Point{X: 0, Y: 1}
	outOfRange := grid.Point{X: 5, Y: 5}

	records := HeightMapUpdateTiles(width, tiles, []grid.Point{pointA, pointB, pointA, outOfRange})

	if len(records) != 2 {
		t.Fatalf("expected 2 deduped records, got %#v", records)
	}
	if records[0].X != 0 || records[0].Y != 0 || records[0].Value != 0 {
		t.Fatalf("unexpected first record %#v", records[0])
	}
	if records[1].X != 0 || records[1].Y != 1 || records[1].Value != int16(512|0x4000) {
		t.Fatalf("unexpected second record %#v", records[1])
	}
}
