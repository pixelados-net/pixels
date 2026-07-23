// Package progressed contains committed room game progression events.
package progressed

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies one committed room game progression delta.
const Name bus.Name = "room.game.progressed"

// Payload describes one player progression delta.
type Payload struct {
	// PlayerID identifies the progressing player.
	PlayerID int64
	// Key identifies the progression trigger.
	Key string
	// Amount stores the positive delta.
	Amount int64
}
