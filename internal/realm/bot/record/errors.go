package record

import "errors"

var (
	// ErrBotNotFound reports a missing bot or ownership mismatch.
	ErrBotNotFound = errors.New("bot not found")
	// ErrRoomNotFound reports a missing active room.
	ErrRoomNotFound = errors.New("bot room not found")
	// ErrNoRights reports bot management without authority.
	ErrNoRights = errors.New("player has no bot rights")
	// ErrRoomLimit reports the configured room bot limit.
	ErrRoomLimit = errors.New("room bot limit reached")
	// ErrInventoryLimit reports the configured inventory bot limit.
	ErrInventoryLimit = errors.New("bot inventory limit reached")
	// ErrTileNotFree reports invalid or occupied placement.
	ErrTileNotFree = errors.New("bot placement tile is not free")
	// ErrInvalidSkill reports malformed bot configuration.
	ErrInvalidSkill = errors.New("invalid bot skill data")
	// ErrConflict reports an optimistic bot update conflict.
	ErrConflict = errors.New("bot update conflict")
	// ErrServeKeywordExists reports a duplicate bartender keyword.
	ErrServeKeywordExists = errors.New("bot serve keyword already exists")
)
