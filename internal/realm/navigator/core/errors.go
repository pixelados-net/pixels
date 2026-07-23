// Package service contains navigator persistence rules.
package core

import "errors"

var (
	// ErrInvalidPlayerID reports a malformed player id.
	ErrInvalidPlayerID = errors.New("invalid navigator player id")

	// ErrInvalidRoomID reports a malformed room id.
	ErrInvalidRoomID = errors.New("invalid navigator room id")

	// ErrInvalidSearchID reports a malformed saved search id.
	ErrInvalidSearchID = errors.New("invalid navigator saved search id")

	// ErrInvalidSearch reports malformed saved search data.
	ErrInvalidSearch = errors.New("invalid navigator saved search")

	// ErrSearchNotFound reports a missing saved search.
	ErrSearchNotFound = errors.New("navigator saved search not found")

	// ErrInvalidPreference reports malformed preference data.
	ErrInvalidPreference = errors.New("invalid navigator preference")
)

const (
	// DefaultWindowX is the default navigator x coordinate.
	DefaultWindowX = 68

	// DefaultWindowY is the default navigator y coordinate.
	DefaultWindowY = 42

	// DefaultWindowWidth is the default navigator width.
	DefaultWindowWidth = 425

	// DefaultWindowHeight is the default navigator height.
	DefaultWindowHeight = 592

	// MaxSearchCodeLength is the maximum saved search code length.
	MaxSearchCodeLength = 64

	// MaxSearchFilterLength is the maximum saved search filter length.
	MaxSearchFilterLength = 128
)
