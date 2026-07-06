// Package unit contains room world unit state primitives.
package unit

import "errors"

var (
	// ErrInvalidUnit reports malformed unit creation input.
	ErrInvalidUnit = errors.New("invalid room unit")
)
