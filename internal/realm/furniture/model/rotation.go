package model

// Rotation stores a floor furniture instance rotation.
type Rotation int16

const (
	// RotationNorth stores north-facing rotation.
	RotationNorth Rotation = 0

	// RotationEast stores east-facing rotation.
	RotationEast Rotation = 2

	// RotationSouth stores south-facing rotation.
	RotationSouth Rotation = 4

	// RotationWest stores west-facing rotation.
	RotationWest Rotation = 6
)

// Valid reports whether the rotation is one of the supported floor values.
func (rotation Rotation) Valid() bool {
	switch rotation {
	case RotationNorth, RotationEast, RotationSouth, RotationWest:
		return true
	default:
		return false
	}
}
