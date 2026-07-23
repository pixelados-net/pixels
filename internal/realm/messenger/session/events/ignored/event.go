// Package ignored contains committed directional ignore events.
package ignored

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies a committed player ignore.
const Name bus.Name = "messenger.player_ignored"

// Payload describes one directional ignore.
type Payload struct {
	// PlayerID identifies the actor.
	PlayerID int64
	// IgnoredPlayerID identifies the hidden target.
	IgnoredPlayerID int64
}
