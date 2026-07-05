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
)
