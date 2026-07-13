package bundle

import "errors"

var (
	// ErrInvalidTemplate reports a missing or unmarked template room.
	ErrInvalidTemplate = errors.New("invalid room bundle template")
	// ErrRoomLimitReached reports a buyer at the room ownership limit.
	ErrRoomLimitReached = errors.New("room ownership limit reached")
	// ErrTemplateReferenced reports a template used by an active offer.
	ErrTemplateReferenced = errors.New("room bundle template is referenced")
	// ErrRoomNotFound reports a missing room administration target.
	ErrRoomNotFound = errors.New("room not found")
)
