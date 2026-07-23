// Package closed contains the navigator closed event.
package closed

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the navigator closed event.
const Name bus.Name = "navigator.closed"

// Payload describes a navigator viewer close event.
type Payload struct {
	// PlayerID identifies the player.
	PlayerID int64
}
