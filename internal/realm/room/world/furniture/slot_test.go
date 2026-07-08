package furniture

import (
	"testing"

	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// TestSlotsReturnsNilWithoutDeclaredSlots verifies blocking items produce no slots.
func TestSlotsReturnsNilWithoutDeclaredSlots(t *testing.T) {
	item := Item{Point: grid.MustPoint(1, 1), Definition: Definition{Width: 2, Length: 2}}
	if slots := Slots(item); slots != nil {
		t.Fatalf("expected no slots, got %#v", slots)
	}
}

// TestSlotsResolvesAbsolutePositionsAtRotationNorth verifies unrotated slot placement.
func TestSlotsResolvesAbsolutePositionsAtRotationNorth(t *testing.T) {
	item := sofaItemForTest(t, worldunit.RotationNorth)

	slots := Slots(item)
	if len(slots) != 2 {
		t.Fatalf("expected two slots, got %#v", slots)
	}
	assertSlot(t, slots[0], grid.MustPoint(1, 1), 1, worldunit.RotationSouth, SlotStatusSit)
	assertSlot(t, slots[1], grid.MustPoint(2, 1), 1, worldunit.RotationSouth, SlotStatusSit)
}

// TestSlotsRotatesOffsetsAndBodyFacingAtRotationEast verifies rotated slot placement.
func TestSlotsRotatesOffsetsAndBodyFacingAtRotationEast(t *testing.T) {
	item := sofaItemForTest(t, worldunit.RotationEast)

	slots := Slots(item)
	if len(slots) != 2 {
		t.Fatalf("expected two slots, got %#v", slots)
	}
	assertSlot(t, slots[0], grid.MustPoint(1, 1), 1, worldunit.RotationWest, SlotStatusSit)
	assertSlot(t, slots[1], grid.MustPoint(1, 2), 1, worldunit.RotationWest, SlotStatusSit)
}

// TestSlotsSkipsOutOfRangePoints verifies malformed offsets are dropped rather than panicking.
func TestSlotsSkipsOutOfRangePoints(t *testing.T) {
	item := Item{
		Point:      grid.MustPoint(0, 0),
		Rotation:   worldunit.RotationNorth,
		Definition: Definition{Width: 1, Length: 1, Slots: []SlotDefinition{{DX: -1, DY: 0, Status: SlotStatusSit}}},
	}

	if slots := Slots(item); len(slots) != 0 {
		t.Fatalf("expected out-of-range slot to be skipped, got %#v", slots)
	}
}

// sofaItemForTest returns a two-slot sofa-shaped item for tests.
func sofaItemForTest(t *testing.T, rotation worldunit.Rotation) Item {
	t.Helper()

	return Item{
		ID:       5,
		Point:    grid.MustPoint(1, 1),
		Z:        0,
		Rotation: rotation,
		Definition: Definition{
			Width: 2, Length: 1, StackHeight: 1, AllowSit: true,
			Slots: []SlotDefinition{
				{DX: 0, DY: 0, Status: SlotStatusSit, BodyRotation: worldunit.RotationSouth},
				{DX: 1, DY: 0, Status: SlotStatusSit, BodyRotation: worldunit.RotationSouth},
			},
		},
	}
}

// assertSlot verifies one resolved slot.
func assertSlot(t *testing.T, slot Slot, point grid.Point, height grid.Height, rotation worldunit.Rotation, status SlotStatus) {
	t.Helper()

	if slot.Point != point || slot.Z != height || slot.BodyRotation != rotation || slot.Status != status {
		t.Fatalf("unexpected slot %#v", slot)
	}
}
