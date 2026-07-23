// Package effectenabled contains the player effect selection event.
package effectenabled

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies a committed effect selection.
const Name bus.Name = "player.effect.enabled"

// Payload describes an enabled or disabled room effect.
type Payload struct {
	// PlayerID identifies the affected player.
	PlayerID int64
	// EffectID identifies the selected effect; zero means disabled.
	EffectID int32
	// Source identifies the effect origin.
	Source string
}
