// Package service contains player application behavior.
package service

import "errors"

var (
	// ErrInvalidPlayerID reports a missing or invalid player identifier.
	ErrInvalidPlayerID = errors.New("invalid player id")

	// ErrPlayerNotFound reports a missing persistent player.
	ErrPlayerNotFound = errors.New("player not found")

	// ErrInvalidUsername reports a missing or invalid username.
	ErrInvalidUsername = errors.New("invalid player username")

	// ErrInvalidLook reports an invalid avatar look.
	ErrInvalidLook = errors.New("invalid player look")

	// ErrInvalidMotto reports an invalid player motto.
	ErrInvalidMotto = errors.New("invalid player motto")

	// ErrInvalidGender reports an invalid profile gender.
	ErrInvalidGender = errors.New("invalid player gender")

	// ErrInvalidHomeRoomID reports a non-positive home room identifier.
	ErrInvalidHomeRoomID = errors.New("invalid player home room id")

	// ErrClubWriterUnavailable reports a player store without club persistence.
	ErrClubWriterUnavailable = errors.New("player club writer unavailable")

	// ErrUsernameTaken reports a player username uniqueness conflict.
	ErrUsernameTaken = errors.New("player username already exists")

	// ErrInvalidBubbleStyle reports a negative Nitro bubble style.
	ErrInvalidBubbleStyle = errors.New("invalid player bubble style")

	// ErrConflict reports a concurrent player mutation.
	ErrConflict = errors.New("player update conflict")

	// ErrAdminWriterUnavailable reports a store without administrative mutations.
	ErrAdminWriterUnavailable = errors.New("player admin writer unavailable")
)
