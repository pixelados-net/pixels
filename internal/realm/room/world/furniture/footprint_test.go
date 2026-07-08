package furniture

import (
	"testing"

	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// TestFootprintUsesRotatedDimensions verifies footprint tile enumeration per rotation.
func TestFootprintUsesRotatedDimensions(t *testing.T) {
	origin := grid.MustPoint(2, 3)

	tests := []struct {
		name     string
		rotation worldunit.Rotation
		expected []grid.Point
	}{
		{
			name:     "north uses width by length",
			rotation: worldunit.RotationNorth,
			expected: []grid.Point{
				grid.MustPoint(2, 3), grid.MustPoint(3, 3),
				grid.MustPoint(2, 4), grid.MustPoint(3, 4),
				grid.MustPoint(2, 5), grid.MustPoint(3, 5),
			},
		},
		{
			name:     "south matches north footprint",
			rotation: worldunit.RotationSouth,
			expected: []grid.Point{
				grid.MustPoint(2, 3), grid.MustPoint(3, 3),
				grid.MustPoint(2, 4), grid.MustPoint(3, 4),
				grid.MustPoint(2, 5), grid.MustPoint(3, 5),
			},
		},
		{
			name:     "east swaps width and length",
			rotation: worldunit.RotationEast,
			expected: []grid.Point{
				grid.MustPoint(2, 3), grid.MustPoint(3, 3), grid.MustPoint(4, 3),
				grid.MustPoint(2, 4), grid.MustPoint(3, 4), grid.MustPoint(4, 4),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			points := Footprint(origin, 2, 3, test.rotation)
			if len(points) != len(test.expected) {
				t.Fatalf("expected %d points, got %d: %#v", len(test.expected), len(points), points)
			}
			for index, point := range test.expected {
				if points[index] != point {
					t.Fatalf("expected point %d to be %#v, got %#v", index, point, points[index])
				}
			}
		})
	}
}

// TestFootprintSingleTile verifies a 1x1 footprint returns its origin only.
func TestFootprintSingleTile(t *testing.T) {
	origin := grid.MustPoint(5, 5)
	points := Footprint(origin, 1, 1, worldunit.RotationNorth)
	if len(points) != 1 || points[0] != origin {
		t.Fatalf("expected single origin point, got %#v", points)
	}
}
