// Package layout contains room layout catalog primitives.
package layout

import "errors"

var (
	// ErrInvalidLayout reports a malformed room layout.
	ErrInvalidLayout = errors.New("invalid room layout")

	// ErrLayoutNotFound reports a missing room layout.
	ErrLayoutNotFound = errors.New("room layout not found")

	// ErrInvalidLayoutID reports a malformed room layout id.
	ErrInvalidLayoutID = errors.New("invalid room layout id")

	// ErrCustomLayoutsUnsupported reports a store without custom layout persistence.
	ErrCustomLayoutsUnsupported = errors.New("custom room layouts are unsupported")
)
