package entry

import "errors"

var (
	// ErrWrongPassword reports a failed room password check.
	ErrWrongPassword = errors.New("wrong room password")
	// ErrEntryLocked reports a temporary password-attempt lockout.
	ErrEntryLocked = errors.New("room entry temporarily locked")
	// ErrAccessDenied reports a room door-mode rejection.
	ErrAccessDenied = errors.New("room access denied")
	// ErrDoorbellRequired reports that owner approval must be requested.
	ErrDoorbellRequired = errors.New("room doorbell approval required")
	// ErrBanned reports a room-specific ban.
	ErrBanned = errors.New("player is banned from room")
	// ErrInvalidPassword reports an invalid password hashing request.
	ErrInvalidPassword = errors.New("invalid room password")
)
