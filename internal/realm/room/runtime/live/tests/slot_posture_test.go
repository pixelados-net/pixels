package live_test

import (
	"testing"

	. "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// TestRoomTickLaysOnTallBedAndWalksOff verifies a real-shaped bed (footprint 1x3, engine height 2,
// pillow-anchored lay slot) is reachable at floor level by clicking any footprint tile, applies the
// lay status with the height offset at the pillow anchor, and does not trap the unit on top after.
func TestRoomTickLaysOnTallBedAndWalksOff(t *testing.T) {
	room := worldRoomForTest(t, "000\r000\r000\r000", 0, 0)
	if _, err := room.Join(occupantForTest(7)); err != nil {
		t.Fatalf("join room: %v", err)
	}
	pillowTile := pointForTest(t, 2, 0)
	footTile := pointForTest(t, 2, 2)
	bed := worldfurniture.Item{
		ID: 6,
		Definition: worldfurniture.Definition{
			Width: 1, Length: 3, StackHeight: grid.HeightFromInt(2), AllowLay: true,
			Slots: []worldfurniture.SlotDefinition{{DX: 0, DY: 0, Status: worldfurniture.SlotStatusLay, BodyRotation: worldunit.RotationNorth}},
		},
		Point: pillowTile, Rotation: worldunit.RotationNorth,
	}
	if _, err := room.ReloadFurniture(6, &bed); err != nil {
		t.Fatalf("place bed: %v", err)
	}

	settleAt(t, room, 7, footTile)

	if playerID, occupied := room.SlotOccupant(6); !occupied || playerID != 7 {
		t.Fatalf("expected player 7 laying on bed, got playerID=%d occupied=%v", playerID, occupied)
	}
	units := room.Units()
	if len(units) != 1 || units[0].Position.Point != pillowTile {
		t.Fatalf("expected foot-tile goal redirected to the pillow anchor, got %#v", units)
	}
	if units[0].BodyRotation != worldunit.RotationNorth {
		t.Fatalf("expected lay facing the bed rotation, got %#v", units)
	}
	if !hasStatusValue(units[0].Statuses, worldunit.StatusLay, "2") {
		t.Fatalf("expected lay status with height offset 2, got %#v", units[0].Statuses)
	}

	settleAt(t, room, 7, pointForTest(t, 0, 3))

	if _, occupied := room.SlotOccupant(6); occupied {
		t.Fatal("expected bed slot released after walking off")
	}
}

// TestRoomFaceToKeepsSettledUnitsInPlace verifies a seated unit ignores facing requests instead of
// spinning or standing when the player clicks an unreachable tile.
func TestRoomFaceToKeepsSettledUnitsInPlace(t *testing.T) {
	room := worldRoomForTest(t, "000", 0, 0)
	if _, err := room.Join(occupantForTest(7)); err != nil {
		t.Fatalf("join room: %v", err)
	}
	seat := pointForTest(t, 2, 0)
	settleOnChair(t, room, 7, 4, seat)

	unit, err := room.FaceTo(7, pointForTest(t, 0, 0))
	if err != nil {
		t.Fatalf("face settled unit: %v", err)
	}
	if !hasStatus(unit.Statuses, worldunit.StatusSit) {
		t.Fatalf("expected sit status kept, got %#v", unit.Statuses)
	}
	if unit.BodyRotation != worldunit.RotationNorth {
		t.Fatalf("expected slot rotation kept, got %#v", unit)
	}
}

// TestRoomTwoKekosShareSofaSlots verifies two units settle independently on a multi-seat sofa.
// The room has a second row because sit tiles are destination-only: the far seat must be reachable
// around the sofa, not through the near seat.
func TestRoomTwoKekosShareSofaSlots(t *testing.T) {
	room := worldRoomForTest(t, "00000\r00000", 0, 0)
	if _, err := room.Join(occupantForTest(7)); err != nil {
		t.Fatalf("join first keko: %v", err)
	}
	if _, err := room.Join(occupantForTest(8)); err != nil {
		t.Fatalf("join second keko: %v", err)
	}

	seatA := pointForTest(t, 2, 0)
	seatB := pointForTest(t, 3, 0)
	sofa := worldfurniture.Item{
		ID: 5,
		Definition: worldfurniture.Definition{
			Width: 2, Length: 1, StackHeight: 0, AllowSit: true,
			Slots: []worldfurniture.SlotDefinition{
				{DX: 0, DY: 0, Status: worldfurniture.SlotStatusSit, BodyRotation: worldunit.RotationNorth},
				{DX: 1, DY: 0, Status: worldfurniture.SlotStatusSit, BodyRotation: worldunit.RotationNorth},
			},
		},
		Point: seatA, Rotation: worldunit.RotationNorth,
	}
	if _, err := room.ReloadFurniture(5, &sofa); err != nil {
		t.Fatalf("place sofa: %v", err)
	}

	settleAt(t, room, 8, seatB)
	settleAt(t, room, 7, seatA)

	unitA, foundA := room.Unit(7)
	unitB, foundB := room.Unit(8)
	if !foundA || !foundB || unitA.Position.Point != seatA || unitB.Position.Point != seatB {
		t.Fatalf("expected independent seat occupants, got unitA=%#v unitB=%#v", unitA, unitB)
	}
	if playerID, occupied := room.SlotOccupant(5); !occupied || (playerID != 7 && playerID != 8) {
		t.Fatalf("expected sofa to report an occupant, got playerID=%d occupied=%v", playerID, occupied)
	}
}

// TestRoomLeaveReleasesSlot verifies leaving the room frees an occupied slot.
func TestRoomLeaveReleasesSlot(t *testing.T) {
	room := worldRoomForTest(t, "000", 0, 0)
	if _, err := room.Join(occupantForTest(7)); err != nil {
		t.Fatalf("join room: %v", err)
	}
	seat := pointForTest(t, 2, 0)
	settleOnChair(t, room, 7, 6, seat)

	if _, found := room.Leave(7); !found {
		t.Fatal("expected leave to succeed")
	}
	if _, occupied := room.SlotOccupant(6); occupied {
		t.Fatal("expected slot released after leave")
	}
}

// TestRoomCloseReleasesAllSlots verifies closing the room frees every occupied slot.
func TestRoomCloseReleasesAllSlots(t *testing.T) {
	room := worldRoomForTest(t, "000", 0, 0)
	if _, err := room.Join(occupantForTest(7)); err != nil {
		t.Fatalf("join room: %v", err)
	}
	seat := pointForTest(t, 2, 0)
	settleOnChair(t, room, 7, 6, seat)

	room.Close()

	if _, occupied := room.SlotOccupant(6); occupied {
		t.Fatal("expected no remaining slot occupants")
	}
}

// TestRoomMoveToWhileSeatedReleasesSlot verifies walking away immediately frees the slot.
func TestRoomMoveToWhileSeatedReleasesSlot(t *testing.T) {
	room := worldRoomForTest(t, "000", 0, 0)
	if _, err := room.Join(occupantForTest(7)); err != nil {
		t.Fatalf("join room: %v", err)
	}
	seat := pointForTest(t, 2, 0)
	settleOnChair(t, room, 7, 6, seat)

	if _, err := room.MoveTo(7, pointForTest(t, 0, 0)); err != nil {
		t.Fatalf("move away: %v", err)
	}
	if _, occupied := room.SlotOccupant(6); occupied {
		t.Fatal("expected slot released immediately on walking away")
	}
}

// TestRoomMoveToOwnSlotTileKeepsSeatedNoOp verifies re-targeting the tile a unit is already settled
// on (e.g. re-clicking the chair you're sitting on) is a no-op: it must not release the slot or clear
// the sit/lay status, since a resolved zero-step path never advances on tick and so would silently
// desync World.slotOccupants from the unit's still-settled status without ever broadcasting the change.
func TestRoomMoveToOwnSlotTileKeepsSeatedNoOp(t *testing.T) {
	room := worldRoomForTest(t, "000", 0, 0)
	if _, err := room.Join(occupantForTest(7)); err != nil {
		t.Fatalf("join room: %v", err)
	}
	seat := pointForTest(t, 2, 0)
	settleOnChair(t, room, 7, 6, seat)

	path, err := room.MoveTo(7, seat)
	if err != nil {
		t.Fatalf("move to own seat: %v", err)
	}
	if len(path.Steps()) != 0 {
		t.Fatalf("expected a trivial path, got %#v", path)
	}

	if playerID, occupied := room.SlotOccupant(6); !occupied || playerID != 7 {
		t.Fatalf("expected player 7 to remain the chair occupant, got playerID=%d occupied=%v", playerID, occupied)
	}
	units := room.Units()
	if len(units) != 1 || !hasStatus(units[0].Statuses, worldunit.StatusSit) {
		t.Fatalf("expected sit status kept after re-clicking the same seat, got %#v", units)
	}
}

// settleOnChair places a one-tile sit chair at point and walks a player onto it until they settle.
func settleOnChair(t *testing.T, room *Room, playerID int64, itemID int64, point grid.Point) {
	t.Helper()

	chair := worldfurniture.Item{
		ID: itemID,
		Definition: worldfurniture.Definition{
			Width: 1, Length: 1, StackHeight: 0, AllowSit: true,
			Slots: []worldfurniture.SlotDefinition{{Status: worldfurniture.SlotStatusSit, BodyRotation: worldunit.RotationNorth}},
		},
		Point: point, Rotation: worldunit.RotationNorth,
	}
	if _, err := room.ReloadFurniture(itemID, &chair); err != nil {
		t.Fatalf("place chair: %v", err)
	}
	settleAt(t, room, playerID, point)
}

// settleAt moves a player toward point and ticks until they settle.
func settleAt(t *testing.T, room *Room, playerID int64, point grid.Point) {
	t.Helper()

	if _, err := room.MoveTo(playerID, point); err != nil {
		t.Fatalf("move to point: %v", err)
	}
	for range 20 {
		for _, movement := range room.Tick() {
			if movement.PlayerID == playerID && movement.Settled {
				return
			}
		}
	}
	t.Fatalf("expected player %d to settle at %#v", playerID, point)
}
