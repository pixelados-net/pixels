package core

import "errors"

var (
	// ErrInvalidRequest reports malformed sanction input.
	ErrInvalidRequest = errors.New("invalid sanction request")
	// ErrUnauthorized reports a missing moderator capability.
	ErrUnauthorized = errors.New("sanction unauthorized")
	// ErrImmune reports a protected target.
	ErrImmune = errors.New("sanction target immune")
	// ErrNotFound reports a missing punishment.
	ErrNotFound = errors.New("punishment not found")
	// ErrApplierExists reports duplicate behavior registration.
	ErrApplierExists = errors.New("sanction applier already registered")
	// ErrLadderEmpty reports missing escalation policy.
	ErrLadderEmpty = errors.New("sanction ladder empty")
)
