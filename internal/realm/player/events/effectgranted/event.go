// Package effectgranted contains the player effect grant event.
package effectgranted

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies a committed player effect grant.
const Name bus.Name = "player.effect.granted"

// Payload describes a committed effect grant.
type Payload struct {
	// PlayerID identifies the receiver.
	PlayerID int64
	// EffectID identifies the granted effect.
	EffectID int32
	// Source identifies the granting capability.
	Source string
}
