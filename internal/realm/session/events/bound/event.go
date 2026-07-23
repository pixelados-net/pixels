// Package bound contains the session bound event.
package bound

import (
	"github.com/niflaot/pixels/internal/realm/session/binding"
	"github.com/niflaot/pixels/pkg/bus"
)

// Name identifies the session bound event.
const Name bus.Name = "session.bound"

// Payload describes a session binding lifecycle event.
type Payload struct {
	// Binding stores the player connection binding.
	Binding binding.Binding
}
