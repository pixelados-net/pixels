package furniture

import (
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// Footprint computes the occupied tiles for a furniture placement.
func Footprint(origin grid.Point, width int, length int, rotation worldunit.Rotation) []grid.Point {
	effectiveWidth, effectiveLength := rotatedDimensions(width, length, rotation)
	points := make([]grid.Point, 0, effectiveWidth*effectiveLength)
	for dy := 0; dy < effectiveLength; dy++ {
		for dx := 0; dx < effectiveWidth; dx++ {
			point, ok := grid.NewPoint(int(origin.X)+dx, int(origin.Y)+dy)
			if !ok {
				continue
			}
			points = append(points, point)
		}
	}

	return points
}

// rotatedDimensions returns the effective width/length after rotation.
func rotatedDimensions(width int, length int, rotation worldunit.Rotation) (int, int) {
	if rotation == worldunit.RotationEast || rotation == worldunit.RotationWest {
		return length, width
	}

	return width, length
}
