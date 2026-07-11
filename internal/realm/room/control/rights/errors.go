// Package rights coordinates persistent room build rights.
package rights

import "errors"

var (
	// ErrInvalidIdentity reports a non-positive room or player id.
	ErrInvalidIdentity = errors.New("invalid room rights identity")
	// ErrRoomNotFound reports a missing room.
	ErrRoomNotFound = errors.New("room rights room not found")
	// ErrAccessDenied reports an actor without rights administration permission.
	ErrAccessDenied = errors.New("room rights access denied")
	// ErrOwnerTarget reports an attempt to grant implicit rights to the owner.
	ErrOwnerTarget = errors.New("room owner already has implicit rights")
)
