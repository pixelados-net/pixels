// Package rightsgranted defines the room rights granted event.
package rightsgranted

import "github.com/niflaot/pixels/pkg/bus"

const (
	// Name identifies a committed room rights grant.
	Name bus.Name = "room.rights_granted"
)

// Payload describes a committed room rights grant.
type Payload struct {
	// RoomID identifies the room.
	RoomID int64
	// PlayerID identifies the rights holder.
	PlayerID int64
	// ActorID identifies the granter.
	ActorID int64
}
