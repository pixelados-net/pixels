// Package identity owns player username availability and committed renames.
package identity

import "errors"

const (
	// ResultAvailable reports an available valid username.
	ResultAvailable int32 = 0
	// ResultTooShort reports a username below the minimum length.
	ResultTooShort int32 = 2
	// ResultTooLong reports a username above the maximum length.
	ResultTooLong int32 = 3
	// ResultInvalid reports unsupported username characters.
	ResultInvalid int32 = 4
	// ResultTaken reports a reserved or existing username.
	ResultTaken int32 = 5
	// ResultDisabled reports that the actor cannot rename.
	ResultDisabled int32 = 6
)

var (
	// ErrReservationMissing reports a commit without a matching short reservation.
	ErrReservationMissing = errors.New("username reservation missing")
	// ErrRenameDisabled reports a durable rename policy rejection.
	ErrRenameDisabled = errors.New("username change disabled")
	// ErrUsernameTaken reports a database uniqueness conflict.
	ErrUsernameTaken = errors.New("username taken")
)

// CheckResult contains username policy and availability output.
type CheckResult struct {
	// Code is the Nitro result code.
	Code int32
	// Username stores the normalized candidate.
	Username string
	// Suggestions stores bounded available alternatives.
	Suggestions []string
}

// RenameResult contains one committed identity replacement.
type RenameResult struct {
	// OldUsername stores the prior visible name.
	OldUsername string
	// NewUsername stores the committed visible name.
	NewUsername string
}
