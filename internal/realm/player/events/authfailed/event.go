// Package authfailed contains the player authentication failed event.
package authfailed

import (
	"github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
)

// Name identifies the player authentication failed event.
const Name bus.Name = "player.authentication_failed"

// Payload describes a player authentication lifecycle event.
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
