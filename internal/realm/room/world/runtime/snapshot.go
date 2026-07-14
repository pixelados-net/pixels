package runtime

import (
	"sort"

	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// Units returns stable unit snapshots in player id order.
func (world *World) Units() []UnitSnapshot {
	playerIDs := world.sortedPlayerIDs()
	units := make([]UnitSnapshot, 0, len(playerIDs))
	for _, playerID := range playerIDs {
		units = append(units, unitSnapshot(playerID, world.units[playerID]))
	}

	return units
}

// Unit returns one unit snapshot without allocating an audience slice.
func (world *World) Unit(playerID int64) (UnitSnapshot, bool) {
	roomUnit, found := world.units[playerID]
	if !found {
		return UnitSnapshot{}, false
	}

	return unitSnapshot(playerID, roomUnit), true
}

// UnitByID returns one unit by its room-local identifier without allocation.
func (world *World) UnitByID(unitID int64) (UnitSnapshot, bool) {
	for playerID, roomUnit := range world.units {
		if roomUnit.ID() == unitID {
			return unitSnapshot(playerID, roomUnit), true
		}
	}
	return UnitSnapshot{}, false
}

// SetUnitStatus stores one status on a player unit.
func (world *World) SetUnitStatus(playerID int64, key string, value string) bool {
	roomUnit, found := world.units[playerID]
	if !found {
		return false
	}
	roomUnit.SetStatus(key, value)

	return true
}

// ClearUnitStatus removes one status from a player unit.
func (world *World) ClearUnitStatus(playerID int64, key string) bool {
	roomUnit, found := world.units[playerID]
	if !found {
		return false
	}
	roomUnit.ClearStatus(key)
	return true
}

// PulseUnitStatus snapshots one temporary status without retaining it in world state.
func (world *World) PulseUnitStatus(playerID int64, key string, value string) (UnitSnapshot, bool) {
	roomUnit, found := world.units[playerID]
	if !found {
		return UnitSnapshot{}, false
	}
	roomUnit.SetStatus(key, value)
	snapshot := unitSnapshot(playerID, roomUnit)
	roomUnit.ClearStatus(key)
	return snapshot, true
}

// FurnitureItems returns stable furniture snapshots in item id order.
func (world *World) FurnitureItems() []worldfurniture.Item {
	ids := make([]int64, 0, len(world.furniture))
	for id := range world.furniture {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(left int, right int) bool {
		return ids[left] < ids[right]
	})
	items := make([]worldfurniture.Item, 0, len(ids))
	for _, id := range ids {
		items = append(items, world.furniture[id])
	}

	return items
}

// FurnitureItem returns one furniture snapshot without allocating.
func (world *World) FurnitureItem(itemID int64) (worldfurniture.Item, bool) {
	item, found := world.furniture[itemID]

	return item, found
}

// InteractionAt returns one interactive item on a tile without allocating.
func (world *World) InteractionAt(point grid.Point) (worldfurniture.Item, bool) {
	itemID, found := world.interactions[point]
	if !found {
		return worldfurniture.Item{}, false
	}
	item, found := world.furniture[itemID]

	return item, found
}

// OtherInteractionAt finds another interaction sharing a tile without allocating.
func (world *World) OtherInteractionAt(point grid.Point, excludedID int64) (worldfurniture.Item, bool) {
	for itemID, item := range world.furniture {
		if itemID != excludedID && item.Point == point && item.Definition.InteractionType != "" && item.Definition.InteractionType != "default" {
			return item, true
		}
	}

	return worldfurniture.Item{}, false
}

// SetFurnitureExtraData updates one runtime furniture visual state.
func (world *World) SetFurnitureExtraData(itemID int64, value string) (worldfurniture.Item, bool) {
	item, found := world.furniture[itemID]
	if !found {
		return worldfurniture.Item{}, false
	}
	item.ExtraData = value
	item.Definition.StackHeight = item.Definition.HeightAtState(value)
	world.furniture[itemID] = item

	return item, true
}

// UpdateFurnitureState atomically changes one item snapshot and its optional fixtures.
func (world *World) UpdateFurnitureState(itemID int64, value string, rebuild bool) (worldfurniture.Item, error) {
	item, found := world.furniture[itemID]
	if !found {
		return worldfurniture.Item{}, ErrFurnitureNotFound
	}
	item.ExtraData = value
	item.Definition.StackHeight = item.Definition.HeightAtState(value)
	if rebuild {
		fixtures, err := worldfurniture.Fixtures(item)
		if err != nil {
			return worldfurniture.Item{}, err
		}
		if err := world.resolver.ReplaceFixtures(itemID, fixtures); err != nil {
			return worldfurniture.Item{}, err
		}
	}
	world.furniture[itemID] = item

	return item, nil
}

// HasUnitInFurnitureFootprint reports whether a unit occupies an item's rotated footprint.
func (world *World) HasUnitInFurnitureFootprint(item worldfurniture.Item) bool {
	width, length := worldfurniture.Dimensions(item.Definition.Width, item.Definition.Length, item.Rotation)
	minX, minY := int(item.Point.X), int(item.Point.Y)
	maxX, maxY := minX+width, minY+length
	for _, roomUnit := range world.units {
		point := roomUnit.Position().Point
		x, y := int(point.X), int(point.Y)
		if x >= minX && x < maxX && y >= minY && y < maxY {
			return true
		}
	}

	return false
}

// SurfaceHeights returns current tile heights in row-major order.
func (world *World) SurfaceHeights() (uint16, uint16, []TileHeight) {
	width, height := world.grid.Width(), world.grid.Height()
	tiles := make([]TileHeight, 0, int(width)*int(height))
	for y := uint16(0); y < height; y++ {
		for x := uint16(0); x < width; x++ {
			section, err := world.resolver.TopSection(grid.Point{X: x, Y: y})
			if err != nil {
				tiles = append(tiles, TileHeight{})
				continue
			}
			tiles = append(tiles, TileHeight{
				Valid: true, Height: section.Z(), StackingBlocked: !section.Stacking(),
			})
		}
	}

	return width, height, tiles
}

// SlotOccupant returns a player occupying one item's sit or lay slot.
func (world *World) SlotOccupant(itemID int64) (int64, bool) {
	item, found := world.furniture[itemID]
	if !found {
		return 0, false
	}
	for _, slot := range worldfurniture.Slots(item) {
		if playerID, occupied := world.slotOccupants[slot.Point]; occupied {
			return playerID, true
		}
	}

	return 0, false
}

// sortedPlayerIDs returns player ids in stable order.
func (world *World) sortedPlayerIDs() []int64 {
	playerIDs := make([]int64, 0, len(world.units))
	for playerID := range world.units {
		playerIDs = append(playerIDs, playerID)
	}
	sort.Slice(playerIDs, func(left int, right int) bool {
		return playerIDs[left] < playerIDs[right]
	})

	return playerIDs
}

// unitSnapshot maps a mutable unit to a stable snapshot.
func unitSnapshot(playerID int64, roomUnit *worldunit.Unit) UnitSnapshot {
	return UnitSnapshot{
		PlayerID: playerID, UnitID: roomUnit.ID(), Position: roomUnit.Position(), Previous: roomUnit.Previous(),
		BodyRotation: roomUnit.BodyRotation(), HeadRotation: roomUnit.HeadRotation(),
		Moving: roomUnit.InMotion(), Statuses: roomUnit.Statuses(), HandItem: roomUnit.HandItem(), Idle: roomUnit.Idle(), IdleSince: roomUnit.IdleSince(), ManualIdle: roomUnit.ManualIdle(), ActiveEffectID: roomUnit.ActiveEffect(),
	}
}
