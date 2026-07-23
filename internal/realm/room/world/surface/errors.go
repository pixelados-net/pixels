// Package surface contains compact room tile column and section resolution.
package surface

import "errors"

var (
	// ErrInvalidFixture reports a malformed room surface fixture.
	ErrInvalidFixture = errors.New("invalid room surface fixture")

	// ErrInvalidTile reports a missing or invalid room grid tile.
	ErrInvalidTile = errors.New("invalid room surface tile")

	// ErrNoSection reports a missing section in a room column.
	ErrNoSection = errors.New("room surface section not found")
)
