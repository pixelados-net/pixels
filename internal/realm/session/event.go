// Package session contains runtime session realm wiring.
package session

import (
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/pkg/bus"
)

const (
	// EventBound reports a player was bound to a live connection.
	EventBound bus.Name = "session.bound"

	// EventUnbound reports a player was unbound from a live connection.
	EventUnbound bus.Name = "session.unbound"
)

// BindingEvent describes a session binding lifecycle event.
type BindingEvent struct {
	// Binding stores the player connection binding.
	Binding binding.Binding
}
