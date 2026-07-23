package service

import (
	"strings"
	"unicode/utf8"

	playermodel "github.com/niflaot/pixels/internal/realm/player/model"
)

const (
	// MinUsernameLength is the minimum player username length.
	MinUsernameLength = 3

	// MaxUsernameLength is the maximum player username length.
	MaxUsernameLength = 15

	// MaxLookLength is the maximum player avatar look length.
	MaxLookLength = 512

	// MaxMottoLength is the maximum player motto length.
	MaxMottoLength = 38
)

// normalizeUsername trims username input.
func normalizeUsername(username string) string {
	return strings.TrimSpace(username)
}

// validateUsername verifies username input.
func validateUsername(username string) error {
	length := utf8.RuneCountInString(username)
	if length < MinUsernameLength || length > MaxUsernameLength {
		return ErrInvalidUsername
	}
	for _, value := range username {
		if !validUsernameRune(value) {
			return ErrInvalidUsername
		}
	}

	return nil
}

// validUsernameRune reports whether one ASCII username character is allowed.
func validUsernameRune(value rune) bool {
	if value >= 'a' && value <= 'z' || value >= 'A' && value <= 'Z' || value >= '0' && value <= '9' {
		return true
	}
	return strings.ContainsRune("_-=!?@:,.'", value)
}

// validatePlayerID verifies player id input.
func validatePlayerID(id int64) error {
	if id <= 0 {
		return ErrInvalidPlayerID
	}

	return nil
}

// validateProfile verifies profile input.
func validateProfile(params CreateProfileParams) error {
	if utf8.RuneCountInString(params.Look) > MaxLookLength {
		return ErrInvalidLook
	}

	if utf8.RuneCountInString(params.Motto) > MaxMottoLength {
		return ErrInvalidMotto
	}

	if !params.Gender.Valid() {
		return ErrInvalidGender
	}

	if params.HomeRoomID != nil && *params.HomeRoomID <= 0 {
		return ErrInvalidHomeRoomID
	}

	return nil
}

// defaultGender returns a supported profile gender.
func defaultGender(gender playermodel.Gender) playermodel.Gender {
	if gender == "" {
		return playermodel.GenderMale
	}

	return gender
}
