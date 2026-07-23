// Package respectgranted defines a committed player respect grant event.
package respectgranted

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies one committed directional player respect grant.
const Name bus.Name = "player.respect.granted"

// Payload describes one respect grant and its recipient.
type Payload struct {
	// ActorPlayerID identifies the player giving respect.
	ActorPlayerID int64
	// TargetPlayerID identifies the player receiving respect.
	TargetPlayerID int64
}
