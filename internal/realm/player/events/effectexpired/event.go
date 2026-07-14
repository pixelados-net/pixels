// Package effectexpired contains the player effect expiry event.
package effectexpired

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies one consumed effect charge.
const Name bus.Name = "player.effect.expired"

// Payload describes a consumed effect charge.
type Payload struct {
	// PlayerID identifies the affected player.
	PlayerID int64
	// EffectID identifies the expired effect.
	EffectID int32
	// RemainingCharges stores the durable charges left.
	RemainingCharges int32
	// Source identifies the effect origin.
	Source string
}
