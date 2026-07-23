package model

// Placement describes required floor coordinates and rotation for placing or moving an item.
type Placement struct {
	// X stores the destination floor tile x coordinate.
	X int

	// Y stores the destination floor tile y coordinate.
	Y int

	// Z stores the resolved destination placement height.
	Z float64

	// Rotation stores the destination floor instance rotation.
	Rotation Rotation
}
