package core

import "errors"

var (
	// ErrDisabled reports disabled call-for-help behavior.
	ErrDisabled = errors.New("moderation reports disabled")
	// ErrThrottled reports an exceeded report window.
	ErrThrottled = errors.New("moderation report throttled")
	// ErrInvalid reports malformed moderation input.
	ErrInvalid = errors.New("invalid moderation request")
	// ErrNotFound reports a missing topic or issue.
	ErrNotFound = errors.New("moderation record not found")
	// ErrPickFailed reports an atomic issue claim lost to another moderator.
	ErrPickFailed = errors.New("moderation issue pick failed")
	// ErrUnauthorized reports missing staff capabilities.
	ErrUnauthorized = errors.New("moderation unauthorized")
)
