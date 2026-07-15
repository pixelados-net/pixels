package runtime

import (
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	"github.com/niflaot/pixels/internal/realm/room/world/surface"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// ResolveFurniturePlacement validates occupancy and stacking for a footprint.
func (world *World) ResolveFurniturePlacement(sourceID int64, footprint []grid.Point) (grid.Height, error) {
	occupied := world.unitPositionsExcludingSource(sourceID)
	var height grid.Height
	for _, point := range footprint {
		if _, blocked := occupied[point]; blocked {
			return 0, ErrTileOccupied
		}
		baseHeight, validTile := world.grid.HeightAt(point)
		if !validTile {
			return 0, ErrInvalidPlacement
		}
		column, err := world.resolver.Column(point)
		if err != nil {
			return 0, ErrInvalidPlacement
		}
		top, ok := topExcludingSource(column, sourceID)
		if !ok {
			if baseHeight > height {
				height = baseHeight
			}
			continue
		}
		if !top.Stacking() {
			return 0, ErrCannotStack
		}
		if top.Top() > height {
			height = top.Top()
		}
	}

	return height, nil
}

// SurfaceColumn returns the resolved vertical column for one tile.
func (world *World) SurfaceColumn(point grid.Point) (surface.Column, error) {
	return world.resolver.Column(point)
}

// nearestWalkableSection resolves a safe anchor near a unit's current height.
func (world *World) nearestWalkableSection(position worldpath.Position) (surface.Section, error) {
	column, err := world.resolver.Column(position.Point)
	if err != nil {
		return surface.Section{}, err
	}
	section, found := column.NearestWalkableSection(position.Z)
	if !found {
		return surface.Section{}, worldpath.ErrInvalidStart
	}

	return section, nil
}

// unitPositionsExcludingSource returns occupied tiles except the source item's seated unit.
func (world *World) unitPositionsExcludingSource(sourceID int64) map[grid.Point]struct{} {
	ownSlots := make(map[grid.Point]struct{})
	if item, ok := world.furniture[sourceID]; ok {
		for _, slot := range worldfurniture.Slots(item) {
			ownSlots[slot.Point] = struct{}{}
		}
	}
	occupied := make(map[grid.Point]struct{}, len(world.units))
	for playerID, roomUnit := range world.units {
		point := roomUnit.Position().Point
		if _, isOwnSlot := ownSlots[point]; isOwnSlot {
			if occupantID, sat := world.slotOccupants[point]; sat && occupantID == playerID {
				continue
			}
		}
		occupied[point] = struct{}{}
	}

	return occupied
}

// topExcludingSource returns the highest section not owned by sourceID.
func topExcludingSource(column surface.Column, sourceID int64) (surface.Section, bool) {
	sections := column.Sections()
	for index := len(sections) - 1; index >= 0; index-- {
		if sections[index].SourceID() != sourceID {
			return sections[index], true
		}
	}

	return surface.Section{}, false
}

// CanChangeFurnitureHeight reports whether no other item is stacked above one item.
func (world *World) CanChangeFurnitureHeight(itemID int64) bool {
	item, found := world.furniture[itemID]
	if !found {
		return false
	}
	for otherID, other := range world.furniture {
		if otherID == itemID || other.Z < item.Top() {
			continue
		}
		if footprintsOverlap(item, other) {
			return false
		}
	}

	return true
}

// ResettleFurnitureUnits updates unit heights over one furniture footprint.
func (world *World) ResettleFurnitureUnits(itemID int64) []UnitSnapshot {
	item, found := world.furniture[itemID]
	if !found {
		return nil
	}
	width, length := worldfurniture.Dimensions(item.Definition.Width, item.Definition.Length, item.Rotation)
	minX, minY := int(item.Point.X), int(item.Point.Y)
	maxX, maxY := minX+width, minY+length
	var snapshots []UnitSnapshot
	for playerID, roomUnit := range world.units {
		point := roomUnit.Position().Point
		x, y := int(point.X), int(point.Y)
		if x < minX || x >= maxX || y < minY || y >= maxY {
			continue
		}
		section, err := world.resolver.TopSection(point)
		if err != nil {
			continue
		}
		roomUnit.SetHeight(section.Z())
		snapshots = append(snapshots, unitSnapshot(playerID, roomUnit))
	}

	return snapshots
}

// footprintsOverlap reports whether two rectangular rotated footprints intersect.
func footprintsOverlap(first worldfurniture.Item, second worldfurniture.Item) bool {
	firstWidth, firstLength := worldfurniture.Dimensions(first.Definition.Width, first.Definition.Length, first.Rotation)
	secondWidth, secondLength := worldfurniture.Dimensions(second.Definition.Width, second.Definition.Length, second.Rotation)
	firstX, firstY := int(first.Point.X), int(first.Point.Y)
	secondX, secondY := int(second.Point.X), int(second.Point.Y)

	return firstX < secondX+secondWidth && firstX+firstWidth > secondX &&
		firstY < secondY+secondLength && firstY+firstLength > secondY
}

// reconcileSlotOccupants updates units affected by one changed furniture item.
func (world *World) reconcileSlotOccupants(previousSlots []worldfurniture.Slot, item *worldfurniture.Item) []UnitSnapshot {
	var updatedSlots []worldfurniture.Slot
	if item != nil {
		updatedSlots = worldfurniture.Slots(*item)
	}
	var affected []UnitSnapshot
	for _, previousSlot := range previousSlots {
		playerID, occupied := world.slotOccupants[previousSlot.Point]
		if !occupied {
			continue
		}
		roomUnit, ok := world.units[playerID]
		if !ok {
			continue
		}
		updatedSlot, found := slotAtPoint(updatedSlots, previousSlot.Point)
		if !found {
			world.releaseSlot(playerID)
			roomUnit.StandUp()
			if section, err := world.resolver.TopSection(previousSlot.Point); err == nil {
				roomUnit.SetHeight(section.Z())
			}
			affected = append(affected, unitSnapshot(playerID, roomUnit))
			continue
		}
		roomUnit.Settle(unitStatusFor(updatedSlot.Status), heightValue(updatedSlot.Z-item.Z), updatedSlot.BodyRotation, updatedSlot.BodyRotation)
		affected = append(affected, unitSnapshot(playerID, roomUnit))
	}

	return affected
}

// occupySlot records a player's current furniture slot.
func (world *World) occupySlot(playerID int64, point grid.Point) {
	world.releaseSlot(playerID)
	world.unitSlots[playerID] = point
	world.slotOccupants[point] = playerID
}

// releaseSlot removes a player's furniture slot reservation.
func (world *World) releaseSlot(playerID int64) {
	point, ok := world.unitSlots[playerID]
	if !ok {
		return
	}
	delete(world.unitSlots, playerID)
	delete(world.slotOccupants, point)
}

// slotAtPoint finds a slot at an exact point.
func slotAtPoint(slots []worldfurniture.Slot, point grid.Point) (worldfurniture.Slot, bool) {
	for _, slot := range slots {
		if slot.Point == point {
			return slot, true
		}
	}

	return worldfurniture.Slot{}, false
}

// unitStatusFor maps a furniture slot status to a unit status.
func unitStatusFor(status worldfurniture.SlotStatus) string {
	if status == worldfurniture.SlotStatusLay {
		return worldunit.StatusLay
	}

	return worldunit.StatusSit
}
