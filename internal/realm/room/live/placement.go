package live

import (
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	"github.com/niflaot/pixels/internal/realm/room/world/surface"
)

// ResolveFurniturePlacement validates a footprint against live occupancy and stacking rules, ignoring
// any existing fixtures owned by sourceID, and returns the resulting placement height.
func (room *Room) ResolveFurniturePlacement(sourceID int64, footprint []grid.Point) (grid.Height, error) {
	room.mutex.RLock()
	defer room.mutex.RUnlock()

	if room.world == nil {
		return 0, ErrWorldNotLoaded
	}

	return room.world.resolvePlacementHeight(footprint, sourceID)
}

// resolvePlacementHeight computes the resulting placement height for a footprint, falling back to the
// tile's plain floor height when excluding sourceID leaves no section (a blocking fixture that replaced
// the tile's base section, per surface.Column.AddSection).
func (world *World) resolvePlacementHeight(footprint []grid.Point, sourceID int64) (grid.Height, error) {
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

// unitPositionsExcludingSource returns the set of tiles currently occupied by units, except a unit
// that is the sit/lay occupant of sourceID's own current slot: that unit will be reconciled by the
// same move/rotate that is resolving this footprint, so it must not block the item's own tile.
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

// topExcludingSource returns the topmost section at a column, ignoring sections owned by sourceID.
func topExcludingSource(column surface.Column, sourceID int64) (surface.Section, bool) {
	sections := column.Sections()
	for index := len(sections) - 1; index >= 0; index-- {
		if sections[index].SourceID() == sourceID {
			continue
		}

		return sections[index], true
	}

	return surface.Section{}, false
}
