package record

import "errors"

var (
	// ErrConflict reports a stale optimistic configuration version.
	ErrConflict = errors.New("WIRED configuration version conflict")
	// ErrItemMissing reports a missing or unplaced WIRED furniture item.
	ErrItemMissing = errors.New("WIRED furniture item is not placed in room")
	// ErrTargetMissing reports a selected furniture item outside the WIRED room.
	ErrTargetMissing = errors.New("WIRED target is not placed in room")
)
