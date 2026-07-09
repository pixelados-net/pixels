package request

import "errors"

var (
	// ErrPlayerNotFound reports a missing live currency holder owner.
	ErrPlayerNotFound = errors.New("currency live player not found")
)
