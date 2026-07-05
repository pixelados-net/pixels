// Package model contains reusable persistent record parts.
package model

import "github.com/google/uuid"

// Identity identifies a durable record.
type Identity struct {
	// ID is the stable record identifier.
	ID uuid.UUID
}

// Empty reports whether the identity has no assigned ID.
func (identity Identity) Empty() bool {
	return identity.ID == uuid.Nil
}
