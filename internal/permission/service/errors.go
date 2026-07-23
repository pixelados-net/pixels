package service

import "errors"

var (
	// ErrInvalidPlayerID reports a non-positive player id.
	ErrInvalidPlayerID = errors.New("invalid permission player id")
	// ErrInvalidGroupID reports a non-positive group id.
	ErrInvalidGroupID = errors.New("invalid permission group id")
	// ErrInvalidGroup reports invalid permission group fields.
	ErrInvalidGroup = errors.New("invalid permission group")
	// ErrInvalidNode reports malformed or unregistered permission syntax.
	ErrInvalidNode = errors.New("invalid permission node")
	// ErrGroupNotFound reports a missing permission group.
	ErrGroupNotFound = errors.New("permission group not found")
	// ErrConflict reports a conflicting permission group mutation.
	ErrConflict = errors.New("permission group update conflict")
	// ErrInheritanceCycle reports cyclic permission group inheritance.
	ErrInheritanceCycle = errors.New("permission group inheritance cycle")
)
