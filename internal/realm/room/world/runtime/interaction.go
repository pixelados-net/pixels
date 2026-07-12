package runtime

import (
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// ApplyControlledInteractionStep moves a unit onto one adjacent interaction tile regardless of normal walkability.
func (world *World) ApplyControlledInteractionStep(playerID int64, point grid.Point, control worldunit.ControlKind) error {
	roomUnit, found := world.units[playerID]
	if !found {
		return ErrUnitNotFound
	}
	if !adjacent(roomUnit.Position().Point, point) {
		return worldpath.ErrInvalidGoal
	}
	section, err := world.resolver.TopSection(point)
	if err != nil {
		return err
	}
	world.releaseSlot(playerID)
	roomUnit.SetPath(worldpath.NewPath([]worldpath.Step{{Position: worldpath.Position{Point: point, Z: section.Z()}}}))
	roomUnit.SetControl(control)

	return nil
}

// adjacent reports whether two points share an edge or corner without being equal.
func adjacent(first grid.Point, second grid.Point) bool {
	dx := int(first.X) - int(second.X)
	if dx < 0 {
		dx = -dx
	}
	dy := int(first.Y) - int(second.Y)
	if dy < 0 {
		dy = -dy
	}

	return (dx != 0 || dy != 0) && dx <= 1 && dy <= 1
}

// SetHandItem replaces one unit's carried hand item and returns its snapshot.
func (world *World) SetHandItem(playerID int64, itemID int32) (UnitSnapshot, error) {
	roomUnit, found := world.units[playerID]
	if !found {
		return UnitSnapshot{}, ErrUnitNotFound
	}
	roomUnit.SetHandItem(itemID)

	return unitSnapshot(playerID, roomUnit), nil
}

// ReleaseControl clears server control and pending movement for one unit.
func (world *World) ReleaseControl(playerID int64) (UnitSnapshot, error) {
	roomUnit, found := world.units[playerID]
	if !found {
		return UnitSnapshot{}, ErrUnitNotFound
	}
	roomUnit.ClearPath()
	roomUnit.SetControl(worldunit.ControlNone)

	return unitSnapshot(playerID, roomUnit), nil
}

// SetUnitControl assigns server control and clears pending movement.
func (world *World) SetUnitControl(playerID int64, control worldunit.ControlKind) (UnitSnapshot, error) {
	roomUnit, found := world.units[playerID]
	if !found {
		return UnitSnapshot{}, ErrUnitNotFound
	}
	roomUnit.ClearPath()
	roomUnit.SetControl(control)

	return unitSnapshot(playerID, roomUnit), nil
}
