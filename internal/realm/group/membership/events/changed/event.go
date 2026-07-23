// Package changed contains the social-group membership changed event.
package changed

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies a committed membership change.
const Name bus.Name = "group.membership.changed"

// Payload identifies the affected membership projection.
type Payload struct {
	// GroupID identifies the affected group.
	GroupID int64
	// PlayerID identifies the affected player.
	PlayerID int64
	// Action identifies the bounded membership mutation.
	Action string
}
