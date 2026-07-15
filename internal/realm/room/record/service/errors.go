// Package service contains room persistence rules.
package service

import "errors"

var (
	// ErrInvalidRoomID reports a malformed room id.
	ErrInvalidRoomID = errors.New("invalid room id")

	// ErrInvalidOwner reports malformed room ownership data.
	ErrInvalidOwner = errors.New("invalid room owner")

	// ErrInvalidRoomName reports a malformed room name.
	ErrInvalidRoomName = errors.New("invalid room name")

	// ErrInvalidDescription reports a malformed room description.
	ErrInvalidDescription = errors.New("invalid room description")

	// ErrInvalidMaxUsers reports a malformed max user count.
	ErrInvalidMaxUsers = errors.New("invalid room max users")

	// ErrInvalidDoorMode reports a malformed door mode.
	ErrInvalidDoorMode = errors.New("invalid room door mode")

	// ErrInvalidTradeMode reports a malformed trade mode.
	ErrInvalidTradeMode = errors.New("invalid room trade mode")

	// ErrInvalidRollerSpeed reports an unsupported roller cycle interval.
	ErrInvalidRollerSpeed = errors.New("invalid room roller speed")

	// ErrRoomNotFound reports a missing room.
	ErrRoomNotFound = errors.New("room not found")

	// ErrLayoutNotAvailable reports a missing or disabled room layout.
	ErrLayoutNotAvailable = errors.New("room layout not available")

	// ErrVersionConflict reports a stale room settings mutation.
	ErrVersionConflict = errors.New("room settings version conflict")

	// ErrPasswordRequired reports password mode without a configured password.
	ErrPasswordRequired = errors.New("room password required")

	// ErrInvalidTag reports malformed room settings tags.
	ErrInvalidTag = errors.New("invalid room tag")

	// ErrReservedTag reports a staff-only room tag.
	ErrReservedTag = errors.New("reserved room tag")

	// ErrProhibitedName reports a globally filtered room name.
	ErrProhibitedName = errors.New("room name contains prohibited text")

	// ErrProhibitedDescription reports a globally filtered room description.
	ErrProhibitedDescription = errors.New("room description contains prohibited text")

	// ErrProhibitedTag reports a globally filtered room tag.
	ErrProhibitedTag = errors.New("room tag contains prohibited text")

	// ErrInvalidChatSettings reports unsupported room chat settings.
	ErrInvalidChatSettings = errors.New("invalid room chat settings")

	// ErrInvalidModerationSettings reports unsupported room moderation settings.
	ErrInvalidModerationSettings = errors.New("invalid room moderation settings")

	// ErrInvalidCategory reports a missing or non-selectable room category.
	ErrInvalidCategory = errors.New("invalid room category")
)

const (
	// MaxRoomNameLength is the maximum room name length.
	MaxRoomNameLength = 25

	// MinRoomNameLength is the minimum room name length.
	MinRoomNameLength = 3

	// MaxRoomDescriptionLength is the maximum room description length.
	MaxRoomDescriptionLength = 128

	// DefaultMaxUsers is the default room capacity.
	DefaultMaxUsers = 25

	// MaxRoomUsers is the maximum supported room capacity.
	MaxRoomUsers = 100

	// MaxRoomsPerPlayer is the room ownership limit before subscriptions exist.
	MaxRoomsPerPlayer = 100

	// MaxRoomTags is the maximum tag count per room.
	MaxRoomTags = 5

	// MaxRoomTagLength is the maximum room tag length.
	MaxRoomTagLength = 32
)
