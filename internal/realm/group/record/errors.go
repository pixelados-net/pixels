package record

import "errors"

var (
	// ErrNotFound reports a missing group resource.
	ErrNotFound = errors.New("social group resource not found")
	// ErrConflict reports an optimistic or uniqueness conflict.
	ErrConflict = errors.New("social group conflict")
	// ErrForbidden reports failed social-role authorization.
	ErrForbidden = errors.New("social group action forbidden")
	// ErrInvalid reports invalid group input.
	ErrInvalid = errors.New("invalid social group input")
	// ErrLimit reports a configured group limit.
	ErrLimit = errors.New("social group limit reached")
	// ErrClosed reports a private, deactivated, or locked resource.
	ErrClosed = errors.New("social group resource closed")
	// ErrAlreadyMember reports duplicate active membership.
	ErrAlreadyMember = errors.New("already a social group member")
	// ErrAlreadyPending reports duplicate pending membership.
	ErrAlreadyPending = errors.New("social group request already pending")
)
