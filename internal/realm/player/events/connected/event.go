// Package connected contains the player connected event.
package connected

import (
	"github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
)

// Name identifies the player connected event.
const Name bus.Name = "player.connected"

// Payload describes a player connection lifecycle event.
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
