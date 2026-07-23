package model

// Base composes common durable record fields.
type Base struct {
	// Identity contains the durable record identifier.
	Identity

	// Timestamps contains durable record timestamps.
	Timestamps

	// SoftDelete contains soft delete state.
	SoftDelete

	// Version contains optimistic locking state.
	Version
}
