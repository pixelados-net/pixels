// Package profileloaded contains the player profile loaded event.
package profileloaded

import (
	"github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
)

// Name identifies the player profile loaded event.
const Name bus.Name = "player.profile_loaded"

// Payload describes a loaded player profile event.
type Payload struct {
	// PlayerID identifies the player when known.
	PlayerID int64

	// ConnectionID identifies the connection.
	ConnectionID connection.ID

	// ConnectionKind identifies the connection family.
	ConnectionKind connection.Kind

	// Reason stores a failure reason when available.
	Reason string
}
