package moderation

import "errors"

var (
	// ErrInvalidIdentity reports a non-positive room or player id.
	ErrInvalidIdentity = errors.New("invalid room moderation identity")
	// ErrRoomNotFound reports a missing room.
	ErrRoomNotFound = errors.New("moderation room not found")
	// ErrAccessDenied reports an unauthorized moderation actor.
	ErrAccessDenied = errors.New("room moderation access denied")
	// ErrTargetProtected reports an unkickable target.
	ErrTargetProtected = errors.New("room moderation target is protected")
	// ErrTargetOwner reports an attempt to moderate the room owner.
	ErrTargetOwner = errors.New("room owner cannot be moderated")
	// ErrSelfTarget reports an actor targeting themselves.
	ErrSelfTarget = errors.New("room moderator cannot target themselves")
	// ErrInvalidMuteDuration reports minutes outside configured limits.
	ErrInvalidMuteDuration = errors.New("invalid room mute duration")
	// ErrInvalidBanDuration reports an unsupported Nitro duration name.
	ErrInvalidBanDuration = errors.New("invalid room ban duration")
)
