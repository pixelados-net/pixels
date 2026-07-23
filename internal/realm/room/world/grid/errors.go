// Package grid contains compact room heightmap grid primitives.
package grid

import "errors"

var (
	// ErrEmptyHeightmap reports an empty room heightmap.
	ErrEmptyHeightmap = errors.New("empty room heightmap")

	// ErrIrregularRows reports rows with different widths.
	ErrIrregularRows = errors.New("irregular room heightmap rows")

	// ErrInvalidHeight reports a heightmap tile that cannot be decoded.
	ErrInvalidHeight = errors.New("invalid room heightmap height")

	// ErrOutOfBounds reports coordinates outside the grid.
	ErrOutOfBounds = errors.New("room grid point out of bounds")

	// ErrInvalidDoor reports a door outside the grid or over an invalid tile.
	ErrInvalidDoor = errors.New("invalid room grid door")
)

// ParseError describes a heightmap parse failure location.
type ParseError struct {
	// Row stores the zero-based row where parsing failed.
	Row int

	// Column stores the zero-based column where parsing failed.
	Column int

	// Rune stores the invalid input rune.
	Rune rune

	// Err stores the underlying parse error.
	Err error
}

// Error returns the parse error message.
func (err ParseError) Error() string {
	return err.Err.Error()
}

// Unwrap returns the underlying parse error.
func (err ParseError) Unwrap() error {
	return err.Err
}
