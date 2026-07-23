package record

import "errors"

var (
	// ErrNotFound reports a missing progression record.
	ErrNotFound = errors.New("progression record not found")
	// ErrConflict reports an optimistic or idempotency conflict.
	ErrConflict = errors.New("progression conflict")
	// ErrInvalid reports malformed progression input.
	ErrInvalid = errors.New("invalid progression input")
	// ErrUnavailable reports a closed campaign or disabled feature.
	ErrUnavailable = errors.New("progression unavailable")
)
