// Package binding maps authenticated players to live connections.
package binding

import (
	"errors"
	"time"

	"github.com/niflaot/pixels/networking/connection"
)

var (
	// ErrInvalidBinding reports an incomplete player connection binding.
	ErrInvalidBinding = errors.New("invalid binding")

	// ErrBindingExists reports a duplicate player or connection binding.
	ErrBindingExists = errors.New("binding exists")

	// ErrBindingNotFound reports a missing player or connection binding.
	ErrBindingNotFound = errors.New("binding not found")
)

// Binding maps one authenticated player to one connection.
type Binding struct {
	// PlayerID identifies the authenticated player.
	PlayerID int64

	// ConnectionID identifies the live connection.
	ConnectionID connection.ID

	// ConnectionKind identifies the connection family.
	ConnectionKind connection.Kind

	// BoundAt stores when the binding was created.
	BoundAt time.Time
}

// Valid reports whether the binding can be registered.
func (binding Binding) Valid() bool {
	return binding.PlayerID > 0 && binding.ConnectionID != "" && binding.ConnectionKind != ""
}

// WithTime returns the binding with a default timestamp when missing.
func (binding Binding) WithTime(now time.Time) Binding {
	if binding.BoundAt.IsZero() {
		binding.BoundAt = now
	}

	return binding
}

// ConnectionKey identifies one connection binding.
type ConnectionKey struct {
	// Kind identifies the connection family.
	Kind connection.Kind

	// ID identifies the connection.
	ID connection.ID
}
