package furniture

import (
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// Slot describes one resolved sit/lay position for a placed furniture item.
type Slot struct {
	// Point stores the absolute tile the slot occupies.
	Point grid.Point

	// Z stores the height a unit rests at when using the slot.
	Z grid.Height

	// BodyRotation stores the forced body facing for a unit using the slot.
	BodyRotation worldunit.Rotation

	// Status describes the slot behavior.
	Status SlotStatus
}

// Slots derives the resolved sit/lay slots for a placed item.
func Slots(item Item) []Slot {
	if len(item.Definition.Slots) == 0 {
		return nil
	}

	top := item.Top()
	slots := make([]Slot, 0, len(item.Definition.Slots))
	for _, declared := range item.Definition.Slots {
		dx, dy := rotateOffset(declared.DX, declared.DY, item.Rotation)
		point, ok := grid.NewPoint(int(item.Point.X)+dx, int(item.Point.Y)+dy)
		if !ok {
			continue
		}
		slots = append(slots, Slot{
			Point:        point,
			Z:            top,
			BodyRotation: rotateBody(declared.BodyRotation, item.Rotation),
			Status:       declared.Status,
		})
	}

	return slots
}

// rotateOffset rotates a local slot offset by the instance rotation.
func rotateOffset(dx int, dy int, rotation worldunit.Rotation) (int, int) {
	if rotation == worldunit.RotationEast || rotation == worldunit.RotationWest {
		return dy, dx
	}

	return dx, dy
}

// rotateBody rotates a declared body facing by the instance rotation.
func rotateBody(declared worldunit.Rotation, rotation worldunit.Rotation) worldunit.Rotation {
	return worldunit.Rotation((int(declared) + int(rotation)) % 8)
}
