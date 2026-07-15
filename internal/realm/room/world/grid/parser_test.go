package grid

import (
	"errors"
	"testing"
)

// TestParseDecodesHeightmap verifies base room heightmap decoding.
func TestParseDecodesHeightmap(t *testing.T) {
	roomGrid, err := Parse("x01\r9az\rXb2", WithDoor(1, 0))
	if err != nil {
		t.Fatalf("parse heightmap: %v", err)
	}

	if roomGrid.Width() != 3 || roomGrid.Height() != 3 {
		t.Fatalf("unexpected dimensions %dx%d", roomGrid.Width(), roomGrid.Height())
	}
	if roomGrid.TileCount() != 9 || roomGrid.ValidCount() != 7 {
		t.Fatalf("unexpected counts total=%d valid=%d", roomGrid.TileCount(), roomGrid.ValidCount())
	}

	assertHeight(t, roomGrid, MustPoint(1, 0), 0)
	assertHeight(t, roomGrid, MustPoint(2, 1), HeightFromInt(35))
	assertHeight(t, roomGrid, MustPoint(1, 2), HeightFromInt(11))
	assertInvalid(t, roomGrid, MustPoint(0, 0))
	assertDoor(t, roomGrid, MustPoint(1, 0))
}

// TestParseNormalizesLineEndings verifies supported row separators.
func TestParseNormalizesLineEndings(t *testing.T) {
	roomGrid, err := Parse("01\n23\r\n45")
	if err != nil {
		t.Fatalf("parse normalized heightmap: %v", err)
	}

	if roomGrid.Width() != 2 || roomGrid.Height() != 3 {
		t.Fatalf("unexpected dimensions %dx%d", roomGrid.Width(), roomGrid.Height())
	}
	assertHeight(t, roomGrid, MustPoint(1, 2), HeightFromInt(5))
}

// TestParseRejectsInvalidInput verifies malformed heightmaps.
func TestParseRejectsInvalidInput(t *testing.T) {
	tests := []struct {
		// name stores the test case name.
		name string

		// heightmap stores the raw heightmap input.
		heightmap string

		// expected stores the expected wrapped error.
		expected error
	}{
		{name: "empty", heightmap: "", expected: ErrEmptyHeightmap},
		{name: "irregular", heightmap: "00\r0", expected: ErrIrregularRows},
		{name: "invalid height", heightmap: "0?", expected: ErrInvalidHeight},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Parse(test.heightmap)
			if !errors.Is(err, test.expected) {
				t.Fatalf("expected %v, got %v", test.expected, err)
			}
		})
	}
}

// TestParseRejectsInvalidDoor verifies door validation.
func TestParseRejectsInvalidDoor(t *testing.T) {
	tests := []struct {
		// name stores the test case name.
		name string

		// x stores the door x coordinate.
		x int

		// y stores the door y coordinate.
		y int
	}{
		{name: "negative", x: -1, y: 0},
		{name: "out of bounds", x: 2, y: 0},
		{name: "invalid tile", x: 0, y: 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Parse("x0", WithDoor(test.x, test.y))
			if !errors.Is(err, ErrInvalidDoor) {
				t.Fatalf("expected invalid door, got %v", err)
			}
		})
	}
}

// assertHeight verifies a tile height.
func assertHeight(t *testing.T, roomGrid Grid, point Point, expected Height) {
	t.Helper()

	height, ok := roomGrid.HeightAt(point)
	if !ok {
		t.Fatalf("expected height at %#v", point)
	}
	if height != expected {
		t.Fatalf("expected height %d at %#v, got %d", expected, point, height)
	}
}

// assertInvalid verifies an invalid tile.
func assertInvalid(t *testing.T, roomGrid Grid, point Point) {
	t.Helper()

	if roomGrid.Valid(point) {
		t.Fatalf("expected invalid tile at %#v", point)
	}
}

// assertDoor verifies a door tile.
func assertDoor(t *testing.T, roomGrid Grid, point Point) {
	t.Helper()

	tile, ok := roomGrid.Tile(point)
	if !ok {
		t.Fatalf("expected tile at %#v", point)
	}
	if !tile.Door() {
		t.Fatalf("expected door at %#v", point)
	}
}
