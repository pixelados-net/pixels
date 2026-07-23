// Package fireworkcharged defines one accepted firework explosion.
package fireworkcharged

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies an accepted firework charge use.
const Name bus.Name = "furniture.firework.charged"

// Payload identifies the actor and furniture instance.
type Payload struct {
	// PlayerID identifies the acting player.
	PlayerID int64
	// ItemID identifies the firework.
	ItemID int64
}
