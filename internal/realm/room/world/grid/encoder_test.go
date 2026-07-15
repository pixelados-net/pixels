package grid

import (
	"errors"
	"testing"
)

// TestEncodeRoundTripsHeightmap verifies normalized heightmap encoding.
func TestEncodeRoundTripsHeightmap(t *testing.T) {
	roomGrid, err := Parse("x01\r9az\rXb2")
	if err != nil {
		t.Fatalf("parse heightmap: %v", err)
	}

	encoded, err := roomGrid.Encode()
	if err != nil {
		t.Fatalf("encode heightmap: %v", err)
	}

	if encoded != "x01\r9az\rxb2" {
		t.Fatalf("unexpected encoded heightmap %q", encoded)
	}
}

// TestEncodeRejectsEmptyGrid verifies empty grid encoding.
func TestEncodeRejectsEmptyGrid(t *testing.T) {
	_, err := Encode(Grid{})
	if !errors.Is(err, ErrEmptyHeightmap) {
		t.Fatalf("expected empty heightmap, got %v", err)
	}
}

// TestEncodeRejectsInvalidHeight verifies encoder height bounds.
func TestEncodeRejectsInvalidHeight(t *testing.T) {
	roomGrid := Grid{
		width:   1,
		height:  1,
		heights: []Height{HeightFromInt(36)},
		flags:   []TileFlag{0},
	}

	_, err := roomGrid.Encode()
	if !errors.Is(err, ErrInvalidHeight) {
		t.Fatalf("expected invalid height, got %v", err)
	}
}
