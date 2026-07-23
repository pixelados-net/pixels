package banzai

import "testing"

// TestTileStep verifies arm, advance, lock, hijack, and locked transitions.
func TestTileStep(t *testing.T) {
	tile := Tile{}
	if points, locked := tile.Step(1, 4, 9); points != 0 || locked || tile.State() != 3 {
		t.Fatalf("unexpected arm: %+v", tile)
	}
	tile.Step(1, 4, 9)
	if points, locked := tile.Step(1, 4, 9); points != 9 || !locked || tile.State() != 5 {
		t.Fatalf("unexpected lock: %+v", tile)
	}
	if points, locked := tile.Step(2, 4, 9); points != 0 || locked || tile.Team != 1 {
		t.Fatal("locked tile was stolen")
	}
	enemy := Tile{Team: 2, Progress: 1}
	if points, _ := enemy.Step(1, 4, 9); points != 4 || enemy.Team != 1 || enemy.Progress != 0 {
		t.Fatalf("unexpected hijack: %+v", enemy)
	}
}

// TestTileStateTable verifies all team state values and hijacks.
func TestTileStateTable(t *testing.T) {
	for team := uint8(1); team <= 4; team++ {
		tile := Tile{}
		want := int(team) * 3
		for progress := 0; progress < 3; progress++ {
			points, locked := tile.Step(team, 4, 9)
			if tile.State() != want+progress || locked != (progress == 2) {
				t.Fatalf("team=%d progress=%d tile=%+v points=%d locked=%t", team, progress, tile, points, locked)
			}
			if progress == 2 && points != 9 || progress != 2 && points != 0 {
				t.Fatalf("team=%d progress=%d points=%d", team, progress, points)
			}
		}
		enemy := team%4 + 1
		tile = Tile{Team: enemy, Progress: 1}
		points, locked := tile.Step(team, 4, 9)
		if points != 4 || locked || tile.Team != team || tile.Progress != 0 {
			t.Fatalf("team=%d hijack=%+v points=%d", team, tile, points)
		}
	}
}

// TestCaptureLargest verifies four-directional enclosure and edge exclusion.
func TestCaptureLargest(t *testing.T) {
	board := NewBoard(7, 7)
	for x := 1; x <= 5; x++ {
		board.Tiles[1*7+x] = Tile{Team: 1}
		board.Tiles[5*7+x] = Tile{Team: 1}
	}
	for y := 1; y <= 5; y++ {
		board.Tiles[y*7+1] = Tile{Team: 1}
		board.Tiles[y*7+5] = Tile{Team: 1}
	}
	captured := board.CaptureLargest(1)
	if len(captured) != 9 {
		t.Fatalf("captured %d tiles, want 9", len(captured))
	}
	for _, index := range captured {
		if !board.Tiles[index].Locked() {
			t.Fatalf("tile %d was not locked", index)
		}
	}
}

// TestCaptureLargestSelectsOnlyTheLargestEnclosure verifies separate arena regions.
func TestCaptureLargestSelectsOnlyTheLargestEnclosure(t *testing.T) {
	board := NewBoard(10, 6)
	for _, point := range [][2]int{{1, 1}, {2, 1}, {3, 1}, {1, 2}, {3, 2}, {1, 3}, {2, 3}, {3, 3}, {5, 1}, {6, 1}, {7, 1}, {8, 1}, {5, 2}, {8, 2}, {5, 3}, {8, 3}, {5, 4}, {6, 4}, {7, 4}, {8, 4}} {
		board.Tiles[point[1]*board.Width+point[0]] = Tile{Team: 1}
	}
	captured := board.CaptureLargest(1)
	if len(captured) != 4 {
		t.Fatalf("captured=%v", captured)
	}
}

// BenchmarkBanzaiFloodFill measures the only allocation-sensitive Banzai algorithm.
func BenchmarkBanzaiFloodFill(b *testing.B) {
	board := NewBoard(16, 16)
	for x := 1; x < 15; x++ {
		board.Tiles[16+x] = Tile{Team: 1}
		board.Tiles[14*16+x] = Tile{Team: 1}
	}
	for y := 1; y < 15; y++ {
		board.Tiles[y*16+1] = Tile{Team: 1}
		board.Tiles[y*16+14] = Tile{Team: 1}
	}
	b.ReportAllocs()
	for range b.N {
		for index := range board.Tiles {
			if board.Tiles[index].Team != 1 {
				board.Tiles[index] = Tile{}
			}
		}
		board.CaptureLargest(1)
	}
}

// TestCompleteRequiresEveryLockedTile verifies empty and partial boards do not finish matches.
func TestCompleteRequiresEveryLockedTile(t *testing.T) {
	board := NewBoard(2, 1)
	if board.Complete() {
		t.Fatal("empty board completed")
	}
	board.Tiles[0] = Tile{Team: 1, Progress: 2}
	if board.Complete() {
		t.Fatal("partial board completed")
	}
	board.Tiles[1] = Tile{Team: 2, Progress: 2}
	if !board.Complete() {
		t.Fatal("locked board did not complete")
	}
}
