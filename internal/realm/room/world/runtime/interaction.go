package runtime

import (
	"sort"
	"strconv"
	"time"

	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// RollUnit repositions one idle unit without taking server movement control.
func (world *World) RollUnit(entityKey int64, position worldpath.Position) (UnitSnapshot, error) {
	roomUnit, found := world.units[entityKey]
	if !found {
		return UnitSnapshot{}, ErrUnitNotFound
	}
	if roomUnit.InMotion() {
		return UnitSnapshot{}, ErrUnitExiting
	}
	roomUnit.Reposition(position, roomUnit.BodyRotation())
	return unitSnapshot(entityKey, roomUnit), nil
}

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
	return world.setUnitIdleAt(playerID, idle, false, at)
}

// SetUnitManualIdleAt replaces one unit's manual AFK projection at one deterministic instant.
func (world *World) SetUnitManualIdleAt(playerID int64, idle bool, at time.Time) (UnitSnapshot, bool) {
	return world.setUnitIdleAt(playerID, idle, idle, at)
}

// setUnitIdleAt replaces one unit's AFK projection and its source.
func (world *World) setUnitIdleAt(playerID int64, idle bool, manual bool, at time.Time) (UnitSnapshot, bool) {
	roomUnit, found := world.units[playerID]
	if !found {
		return UnitSnapshot{}, false
	}
	if manual {
		roomUnit.SetManualIdleAt(idle, at)
	} else {
		roomUnit.SetIdleAt(idle, at)
	}
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

// FurnitureByInteraction returns stable furniture snapshots for one indexed interaction type.
func (world *World) FurnitureByInteraction(interactionType string) []worldfurniture.Item {
	ids := world.interactionTypes[interactionType]
	if len(ids) == 0 {
		return nil
	}
	items := make([]worldfurniture.Item, 0, len(ids))
	for _, id := range ids {
		if item, found := world.furniture[id]; found {
			items = append(items, item)
		}
	}
	sort.Slice(items, func(left int, right int) bool { return items[left].ID < items[right].ID })
	return items
}

// addFurnitureIndexes adds one item's footprint and interaction type.
func (world *World) addFurnitureIndexes(item worldfurniture.Item) {
	if world.furnitureTiles == nil {
		world.furnitureTiles = make(map[grid.Point][]int64)
	}
	for _, point := range worldfurniture.Footprint(item.Point, item.Definition.Width, item.Definition.Length, item.Rotation) {
		world.furnitureTiles[point] = append(world.furnitureTiles[point], item.ID)
	}
	if item.Definition.InteractionType == "" {
		return
	}
	if world.interactionTypes == nil {
		world.interactionTypes = make(map[string][]int64)
	}
	world.interactionTypes[item.Definition.InteractionType] = append(world.interactionTypes[item.Definition.InteractionType], item.ID)
}

// removeFurnitureIndexes removes one item's footprint and interaction type.
func (world *World) removeFurnitureIndexes(item worldfurniture.Item) {
	for _, point := range worldfurniture.Footprint(item.Point, item.Definition.Width, item.Definition.Length, item.Rotation) {
		world.furnitureTiles[point] = withoutID(world.furnitureTiles[point], item.ID)
		if len(world.furnitureTiles[point]) == 0 {
			delete(world.furnitureTiles, point)
		}
	}
	kind := item.Definition.InteractionType
	world.interactionTypes[kind] = withoutID(world.interactionTypes[kind], item.ID)
	if len(world.interactionTypes[kind]) == 0 {
		delete(world.interactionTypes, kind)
	}
}

// withoutID removes one id while preserving stable order.
func withoutID(ids []int64, removed int64) []int64 {
	for index, id := range ids {
		if id == removed {
			return append(ids[:index], ids[index+1:]...)
		}
	}
	return ids
}

// interactive reports whether an item participates in movement interaction events.
func interactive(item worldfurniture.Item) bool {
	return item.Definition.InteractionType != "" &&
		(item.Definition.InteractionType != "default" || item.Definition.InteractionModesCount > 1)
}
