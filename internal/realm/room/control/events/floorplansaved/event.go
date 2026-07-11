// Package floorplansaved defines the committed room floor plan event.
package floorplansaved

import "github.com/niflaot/pixels/pkg/bus"

const (
	// Name identifies a committed room floor plan mutation.
	Name bus.Name = "room.floorplan_saved"
)

// Payload contains committed floor plan event data.
type Payload struct {
	// RoomID identifies the updated room.
	RoomID int64
	// ActorID identifies the player that saved the floor plan.
	ActorID int64
}
