package live_test

import (
	"testing"

	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// TestRoomTickSettlesOnFurnitureSlotRotation verifies settling forces the slot's declared rotation.
func TestRoomTickSettlesOnFurnitureSlotRotation(t *testing.T) {
	room := worldRoomForTest(t, "000", 0, 0)
	if _, err := room.Join(occupantForTest(7)); err != nil {
		t.Fatalf("join room: %v", err)
	}
	seat := pointForTest(t, 2, 0)
	chair := worldfurniture.Item{
		ID: 4,
		Definition: worldfurniture.Definition{
			Width: 1, Length: 1, StackHeight: 0, AllowSit: true,
			Slots: []worldfurniture.SlotDefinition{{Status: worldfurniture.SlotStatusSit, BodyRotation: worldunit.RotationWest}},
		},
		Point: seat, Rotation: worldunit.RotationNorth,
	}
	if _, err := room.ReloadFurniture(4, &chair); err != nil {
		t.Fatalf("place chair: %v", err)
	}

	settleAt(t, room, 7, seat)

	playerID, occupied := room.SlotOccupant(4)
	if !occupied || playerID != 7 {
		t.Fatalf("expected player 7 to occupy the chair, got playerID=%d occupied=%v", playerID, occupied)
	}
	units := room.Units()
	if len(units) != 1 || units[0].BodyRotation != worldunit.RotationWest || units[0].HeadRotation != worldunit.RotationWest {
		t.Fatalf("expected forced slot rotation, got %#v", units)
	}
	if !hasStatusValue(units[0].Statuses, worldunit.StatusSit, "0") {
		t.Fatalf("expected sit status at height 0, got %#v", units[0].Statuses)
	}
}

// TestReloadFurnitureRotatesOccupantInPlace verifies rotating a chair with the same anchor tile
// re-settles its occupant with the new forced rotation instead of standing them up.
func TestReloadFurnitureRotatesOccupantInPlace(t *testing.T) {
	room := worldRoomForTest(t, "000", 0, 0)
	if _, err := room.Join(occupantForTest(7)); err != nil {
		t.Fatalf("join room: %v", err)
	}
	seat := pointForTest(t, 2, 0)
	chair := worldfurniture.Item{
		ID: 4,
		Definition: worldfurniture.Definition{
			Width: 1, Length: 1, StackHeight: 0, AllowSit: true,
			Slots: []worldfurniture.SlotDefinition{{Status: worldfurniture.SlotStatusSit, BodyRotation: worldunit.RotationNorth}},
		},
		Point: seat, Rotation: worldunit.RotationNorth,
	}
	if _, err := room.ReloadFurniture(4, &chair); err != nil {
		t.Fatalf("place chair: %v", err)
	}
	settleAt(t, room, 7, seat)

	rotated := chair
	rotated.Rotation = worldunit.RotationEast
	affected, err := room.ReloadFurniture(4, &rotated)
	if err != nil {
		t.Fatalf("rotate chair: %v", err)
	}

	if len(affected) != 1 || affected[0].PlayerID != 7 || affected[0].BodyRotation != worldunit.RotationEast {
		t.Fatalf("expected occupant reoriented east, got %#v", affected)
	}
	if _, occupied := room.SlotOccupant(4); !occupied {
		t.Fatal("expected occupant to remain seated after rotating in place")
	}
	units := room.Units()
	if len(units) != 1 || !hasStatus(units[0].Statuses, worldunit.StatusSit) {
		t.Fatalf("expected sit status kept after rotation, got %#v", units)
	}
}

// TestReloadFurnitureMovingAwayStandsOccupantUp verifies moving a chair to a different tile stands
// its occupant up and releases the slot.
func TestReloadFurnitureMovingAwayStandsOccupantUp(t *testing.T) {
	room := worldRoomForTest(t, "000", 0, 0)
	if _, err := room.Join(occupantForTest(7)); err != nil {
		t.Fatalf("join room: %v", err)
	}
	seat := pointForTest(t, 2, 0)
	chair := worldfurniture.Item{
		ID: 4,
		Definition: worldfurniture.Definition{
			Width: 1, Length: 1, StackHeight: 0, AllowSit: true,
			Slots: []worldfurniture.SlotDefinition{{Status: worldfurniture.SlotStatusSit, BodyRotation: worldunit.RotationNorth}},
		},
		Point: seat, Rotation: worldunit.RotationNorth,
	}
	if _, err := room.ReloadFurniture(4, &chair); err != nil {
		t.Fatalf("place chair: %v", err)
	}
	settleAt(t, room, 7, seat)

	moved := chair
	moved.Point = pointForTest(t, 1, 0)
	affected, err := room.ReloadFurniture(4, &moved)
	if err != nil {
		t.Fatalf("move chair: %v", err)
	}

	if len(affected) != 1 || affected[0].PlayerID != 7 || hasStatus(affected[0].Statuses, worldunit.StatusSit) {
		t.Fatalf("expected occupant standing up, got %#v", affected)
	}
	if _, occupied := room.SlotOccupant(4); occupied {
		t.Fatal("expected slot released after the chair moved away")
	}
}

// TestReloadFurnitureMovingAwayFromElevatedSlotFixesStaleHeight verifies standing up from a slot
// whose furniture was itself sitting on an elevated base (e.g. a real bed_silo_one placed on top of
// something else, resolving item.Z above the bare floor) corrects the unit's stale elevated Z back
// down to the tile's current height once that furniture moves away, so a later walk command
// validates against where the unit actually stands instead of erroring with ErrInvalidStart — the
// exact failure that force-disconnected a real client after standing up from a moved bed.
func TestReloadFurnitureMovingAwayFromElevatedSlotFixesStaleHeight(t *testing.T) {
	room := worldRoomForTest(t, "000\r000", 0, 0)
	if _, err := room.Join(occupantForTest(7)); err != nil {
		t.Fatalf("join room: %v", err)
	}
	spot := pointForTest(t, 2, 0)
	bed := worldfurniture.Item{
		ID: 4,
		Definition: worldfurniture.Definition{
			Width: 1, Length: 1, StackHeight: 2, AllowLay: true,
			Slots: []worldfurniture.SlotDefinition{{Status: worldfurniture.SlotStatusLay, BodyRotation: worldunit.RotationNorth}},
		},
		Point: spot, Z: 1, Rotation: worldunit.RotationNorth,
	}
	if _, err := room.ReloadFurniture(4, &bed); err != nil {
		t.Fatalf("place bed: %v", err)
	}
	settleAt(t, room, 7, spot)

	units := room.Units()
	if len(units) != 1 || units[0].Position.Z != 1 {
		t.Fatalf("expected unit settled at the bed's elevated base height, got %#v", units)
	}

	moved := bed
	moved.Point = pointForTest(t, 1, 0)
	moved.Z = 0
	affected, err := room.ReloadFurniture(4, &moved)
	if err != nil {
		t.Fatalf("move bed: %v", err)
	}

	if len(affected) != 1 || affected[0].Position.Z != 0 {
		t.Fatalf("expected occupant's stale elevated height corrected to the bare floor, got %#v", affected)
	}
	if _, err := room.MoveTo(7, pointForTest(t, 2, 1)); err != nil {
		t.Fatalf("expected walk after standing up to validate cleanly, got %v", err)
	}
}

// TestReloadFurniturePickupStandsOccupantUp verifies picking up a chair (item nil) stands its
// occupant up and releases the slot.
func TestReloadFurniturePickupStandsOccupantUp(t *testing.T) {
	room := worldRoomForTest(t, "000", 0, 0)
	if _, err := room.Join(occupantForTest(7)); err != nil {
		t.Fatalf("join room: %v", err)
	}
	seat := pointForTest(t, 2, 0)
	chair := worldfurniture.Item{
		ID: 4,
		Definition: worldfurniture.Definition{
			Width: 1, Length: 1, StackHeight: 0, AllowSit: true,
			Slots: []worldfurniture.SlotDefinition{{Status: worldfurniture.SlotStatusSit, BodyRotation: worldunit.RotationNorth}},
		},
		Point: seat, Rotation: worldunit.RotationNorth,
	}
	if _, err := room.ReloadFurniture(4, &chair); err != nil {
		t.Fatalf("place chair: %v", err)
	}
	settleAt(t, room, 7, seat)

	affected, err := room.ReloadFurniture(4, nil)
	if err != nil {
		t.Fatalf("pick up chair: %v", err)
	}

	if len(affected) != 1 || affected[0].PlayerID != 7 || hasStatus(affected[0].Statuses, worldunit.StatusSit) {
		t.Fatalf("expected occupant standing up, got %#v", affected)
	}
	if _, occupied := room.SlotOccupant(4); occupied {
		t.Fatal("expected slot released after pickup")
	}
}

// TestRoomTickSettlesOnRaisedSlotNotTiedToBaseFloor verifies settling still lands on the slot when
// its height does not tie with the base floor (e.g. a real chair_plasto with stack_height=1).
func TestRoomTickSettlesOnRaisedSlotNotTiedToBaseFloor(t *testing.T) {
	room := worldRoomForTest(t, "000", 0, 0)
	if _, err := room.Join(occupantForTest(7)); err != nil {
		t.Fatalf("join room: %v", err)
	}
	seat := pointForTest(t, 2, 0)
	chair := worldfurniture.Item{
		ID: 4,
		Definition: worldfurniture.Definition{
			Width: 1, Length: 1, StackHeight: 1, AllowSit: true,
			Slots: []worldfurniture.SlotDefinition{{Status: worldfurniture.SlotStatusSit, BodyRotation: worldunit.RotationWest}},
		},
		Point: seat, Rotation: worldunit.RotationNorth,
	}
	if _, err := room.ReloadFurniture(4, &chair); err != nil {
		t.Fatalf("place chair: %v", err)
	}

	settleAt(t, room, 7, seat)

	if _, occupied := room.SlotOccupant(4); !occupied {
		t.Fatal("expected player to settle onto the raised slot")
	}
}
