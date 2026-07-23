// Package updated defines a committed public player profile update event.
package updated

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies a committed public profile change.
const Name bus.Name = "player.profile.updated"

// Payload describes one committed public profile change.
type Payload struct {
	// PlayerID identifies the changed player.
	PlayerID int64
	// Figure reports an avatar appearance change.
	Figure bool
	// Motto reports a public motto change.
	Motto bool
}
