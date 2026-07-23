// Package unbound contains the session unbound event.
package unbound

import (
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/pkg/bus"
)

// Name identifies the session unbound event.
const Name bus.Name = "session.unbound"

// Payload describes a session unbinding lifecycle event.
type Payload struct {
	// Binding stores the player connection binding.
	Binding binding.Binding
}
