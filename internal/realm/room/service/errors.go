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

	// ErrRoomNotFound reports a missing room.
	ErrRoomNotFound = errors.New("room not found")

	// ErrLayoutNotAvailable reports a missing or disabled room layout.
	ErrLayoutNotAvailable = errors.New("room layout not available")
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

	// MaxRoomTags is the maximum tag count per room.
	MaxRoomTags = 5
)
