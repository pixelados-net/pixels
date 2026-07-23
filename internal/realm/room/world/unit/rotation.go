package unit

import "github.com/niflaot/pixels/internal/realm/room/world/grid"

// Rotation stores a Habbo-style directional rotation.
type Rotation uint8

const (
	// RotationNorth stores north-facing rotation.
	RotationNorth Rotation = 0

	// RotationNorthEast stores north-east-facing rotation.
	RotationNorthEast Rotation = 1

	// RotationEast stores east-facing rotation.
	RotationEast Rotation = 2

	// RotationSouthEast stores south-east-facing rotation.
	RotationSouthEast Rotation = 3

	// RotationSouth stores south-facing rotation.
	RotationSouth Rotation = 4

	// RotationSouthWest stores south-west-facing rotation.
	RotationSouthWest Rotation = 5

	// RotationWest stores west-facing rotation.
	RotationWest Rotation = 6

	// RotationNorthWest stores north-west-facing rotation.
	RotationNorthWest Rotation = 7
)

// RotationBetween returns the rotation from one point to another.
func RotationBetween(fromX uint16, fromY uint16, toX uint16, toY uint16) Rotation {
	dx := compareUint16(toX, fromX)
	dy := compareUint16(toY, fromY)
	switch {
	case dx == 0 && dy < 0:
		return RotationNorth
	case dx > 0 && dy < 0:
		return RotationNorthEast
	case dx > 0 && dy == 0:
		return RotationEast
	case dx > 0 && dy > 0:
		return RotationSouthEast
	case dx == 0 && dy > 0:
		return RotationSouth
	case dx < 0 && dy > 0:
		return RotationSouthWest
	case dx < 0 && dy == 0:
		return RotationWest
	case dx < 0 && dy < 0:
		return RotationNorthWest
	default:
		return RotationSouth
	}
}

// FaceToward rotates the unit toward a target point.
func (unit *Unit) FaceToward(point grid.Point) {
	rotation := RotationBetween(unit.position.Point.X, unit.position.Point.Y, point.X, point.Y)
	unit.body = rotation
	unit.head = rotation
}

// compareUint16 compares two unsigned coordinates.
func compareUint16(left uint16, right uint16) int {
	if left > right {
		return 1
	}
	if left < right {
		return -1
	}

	return 0
}
