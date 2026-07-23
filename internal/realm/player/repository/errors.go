package repository

import "errors"

var (
	// ErrInvalidGender reports an unsupported avatar gender code.
	ErrInvalidGender = errors.New("invalid player profile gender")

	// ErrUsernameTaken reports a player username uniqueness conflict.
	ErrUsernameTaken = errors.New("player username already exists")
)
