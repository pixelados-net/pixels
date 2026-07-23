// Package disconnected contains the player disconnected event.
package disconnected

import (
	"github.com/niflaot/pixels/networking/connection"
	"github.com/niflaot/pixels/pkg/bus"
)

// Name identifies the player disconnected event.
const Name bus.Name = "player.disconnected"

// Payload describes a player disconnection event.
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
