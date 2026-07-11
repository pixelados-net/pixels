// Package furniture computes room world footprints, slots, and surface fixtures for placed furniture.
package furniture

import (
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// SlotStatus describes what a resolved slot does to a unit standing on it.
type SlotStatus string

const (
	// SlotStatusSit marks a slot that seats a unit.
	SlotStatusSit SlotStatus = "sit"

	// SlotStatusLay marks a slot that lays a unit down.
	SlotStatusLay SlotStatus = "lay"
)

// SlotDefinition describes one static sit/lay slot in unrotated local footprint coordinates.
type SlotDefinition struct {
	// DX stores the local x offset within the footprint at rotation 0.
	DX int

	// DY stores the local y offset within the footprint at rotation 0.
	DY int

	// Status describes the slot behavior.
	Status SlotStatus

	// BodyRotation stores the forced body facing at rotation 0.
	BodyRotation worldunit.Rotation
}

// Definition stores the minimal furniture definition snapshot used by the room world.
type Definition struct {
	// SpriteID stores the Nitro rendering class id.
	SpriteID int

	// InteractionType identifies the furniture behavior.
	InteractionType string

	// Width stores the footprint width at rotation 0.
	Width int

	// Length stores the footprint length at rotation 0.
	Length int

	// StackHeight stores the height the definition adds above what it sits on.
	StackHeight grid.Height

	// AllowStack reports whether other items can stack on top.
	AllowStack bool

	// AllowWalk reports whether units can walk over the definition outside its slots.
	AllowWalk bool

	// AllowSit reports whether the definition produces sit slots.
	AllowSit bool

	// AllowLay reports whether the definition produces lay slots.
	AllowLay bool

	// Slots stores declared sit/lay slots in unrotated local coordinates.
	Slots []SlotDefinition
}
