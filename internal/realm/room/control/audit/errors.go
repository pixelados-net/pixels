// Package audit stores and queries append-only room history.
package audit

import "errors"

var (
	// ErrInvalidQuery reports malformed audit filters.
	ErrInvalidQuery = errors.New("invalid room audit query")
)
