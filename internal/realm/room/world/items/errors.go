package furniture

import "errors"

var (
	// ErrInvalidTarget reports a malformed placement target (bad rotation or out-of-grid tile).
	ErrInvalidTarget = errors.New("invalid furniture placement target")

	// ErrDefinitionNotFound reports a missing furniture definition for an item.
	ErrDefinitionNotFound = errors.New("furniture definition not found")
)
