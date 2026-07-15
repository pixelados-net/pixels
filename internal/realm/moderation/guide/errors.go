package guide

import "errors"

var (
	// ErrUnavailable reports no eligible guide or active session.
	ErrUnavailable = errors.New("guide unavailable")
	// ErrBusy reports a duplicate session participant.
	ErrBusy = errors.New("guide participant busy")
	// ErrInvalidState reports an action outside session lifecycle.
	ErrInvalidState = errors.New("invalid guide session state")
	// ErrUnauthorized reports an action by the wrong participant.
	ErrUnauthorized = errors.New("guide session unauthorized")
)
