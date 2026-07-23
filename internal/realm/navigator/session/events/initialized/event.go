// Package initialized contains the navigator initialized event.
package initialized

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies the navigator initialized event.
const Name bus.Name = "navigator.initialized"

// Payload describes a navigator viewer lifecycle event.
type Payload struct {
	// PlayerID identifies the player.
	PlayerID int64
}
