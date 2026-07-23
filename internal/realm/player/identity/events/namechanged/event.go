// Package namechanged defines the committed player identity change event.
package namechanged

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies a committed player username change.
const Name bus.Name = "player.identity.name_changed"

// Payload describes one committed username replacement.
type Payload struct {
	// PlayerID identifies the renamed player.
	PlayerID int64
	// OldUsername stores the prior visible username.
	OldUsername string
	// NewUsername stores the committed visible username.
	NewUsername string
}
