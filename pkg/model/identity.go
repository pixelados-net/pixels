// Package model contains reusable persistent record parts.
package model

// Identity identifies a durable record.
type Identity struct {
	// ID is the stable record identifier.
	ID int64
}

// Empty reports whether the identity has no assigned ID.
func (identity Identity) Empty() bool {
	return identity.ID == 0
}
