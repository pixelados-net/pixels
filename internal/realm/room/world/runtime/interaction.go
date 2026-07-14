package runtime

import (
	"strconv"
	"time"

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

// SetUnitIdle replaces one unit's AFK projection.
func (world *World) SetUnitIdle(playerID int64, idle bool) (UnitSnapshot, bool) {
	return world.SetUnitIdleAt(playerID, idle, time.Now())
}

// SetUnitIdleAt replaces one unit's AFK projection at one deterministic instant.
func (world *World) SetUnitIdleAt(playerID int64, idle bool, at time.Time) (UnitSnapshot, bool) {
	roomUnit, found := world.units[playerID]
	if !found {
		return UnitSnapshot{}, false
	}
	roomUnit.SetIdleAt(idle, at)
	if idle {
		roomUnit.ClearStatus(worldunit.StatusDance)
	}
	return unitSnapshot(playerID, roomUnit), true
}

// SetUnitPosture changes one unit's free-standing posture.
func (world *World) SetUnitPosture(playerID int64, sitting bool) (UnitSnapshot, bool) {
	roomUnit, found := world.units[playerID]
	if !found || roomUnit.Moving() {
		return UnitSnapshot{}, false
	}
	world.releaseSlot(playerID)
	roomUnit.SetFloorPosture(sitting)
	return unitSnapshot(playerID, roomUnit), true
}

// SetUnitDance changes one unit's persistent dance state.
func (world *World) SetUnitDance(playerID int64, danceID int32) (UnitSnapshot, bool) {
	roomUnit, found := world.units[playerID]
	if !found {
		return UnitSnapshot{}, false
	}
	world.releaseSlot(playerID)
	roomUnit.StandUp()
	roomUnit.ClearStatus(worldunit.StatusDance)
	if danceID > 0 {
		roomUnit.SetStatus(worldunit.StatusDance, strconv.FormatInt(int64(danceID), 10))
	}
	return unitSnapshot(playerID, roomUnit), true
}

// SetUnitEffect replaces one unit's selected avatar effect.
func (world *World) SetUnitEffect(playerID int64, effectID int32) (UnitSnapshot, bool) {
	roomUnit, found := world.units[playerID]
	if !found {
		return UnitSnapshot{}, false
	}
	roomUnit.SetActiveEffect(effectID)
	return unitSnapshot(playerID, roomUnit), true
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

// StopMovement discards future steps and lets the current client projection settle.
func (world *World) StopMovement(playerID int64) (bool, error) {
	roomUnit, found := world.units[playerID]
	if !found {
		return false, ErrUnitNotFound
	}

	return roomUnit.StopMovement(), nil
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
