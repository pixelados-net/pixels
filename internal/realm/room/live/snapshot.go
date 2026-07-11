package live

import (
	"sort"

	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// TileHeight describes one resolved tile's current walkable height and stacking state.
type TileHeight struct {
	// Valid reports whether the tile is part of the room.
	Valid bool
	// Height stores the current walkable top height.
	Height grid.Height
	// StackingBlocked reports whether new items cannot stack on this tile.
	StackingBlocked bool
}

// Units returns stable world unit snapshots.
func (room *Room) Units() []UnitSnapshot {
	room.mutex.RLock()
	defer room.mutex.RUnlock()
	if room.world == nil {
		return nil
	}
	playerIDs := make([]int64, 0, len(room.world.units))
	for playerID := range room.world.units {
		playerIDs = append(playerIDs, playerID)
	}
	sort.Slice(playerIDs, func(left int, right int) bool {
		return playerIDs[left] < playerIDs[right]
	})
	units := make([]UnitSnapshot, 0, len(playerIDs))
	for _, playerID := range playerIDs {
		units = append(units, unitSnapshot(playerID, room.world.units[playerID]))
	}

	return units
}

// Presences returns occupant and unit snapshots with one allocation and one lock.
func (room *Room) Presences() []Presence {
	room.mutex.RLock()
	defer room.mutex.RUnlock()
	if room.world == nil {
		return nil
	}
	presences := make([]Presence, 0, len(room.occupants))
	for playerID, occupant := range room.occupants {
		roomUnit, found := room.world.units[playerID]
		if !found {
			continue
		}
		presences = append(presences, Presence{Occupant: occupant, Unit: unitSnapshot(playerID, roomUnit)})
	}

	return presences
}

// Unit returns one unit snapshot without allocating an audience slice.
func (room *Room) Unit(playerID int64) (UnitSnapshot, bool) {
	room.mutex.RLock()
	defer room.mutex.RUnlock()
	if room.world == nil {
		return UnitSnapshot{}, false
	}
	roomUnit, found := room.world.units[playerID]
	if !found {
		return UnitSnapshot{}, false
	}

	return unitSnapshot(playerID, roomUnit), true
}

// SetUnitStatus stores one status on a player unit when its world is loaded.
func (room *Room) SetUnitStatus(playerID int64, key string, value string) bool {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	if room.world == nil {
		return false
	}
	roomUnit, found := room.world.units[playerID]
	if !found {
		return false
	}
	roomUnit.SetStatus(key, value)

	return true
}

// FurnitureItems returns stable placed furniture item snapshots.
func (room *Room) FurnitureItems() []worldfurniture.Item {
	room.mutex.RLock()
	defer room.mutex.RUnlock()
	if room.world == nil {
		return nil
	}
	ids := make([]int64, 0, len(room.world.furniture))
	for id := range room.world.furniture {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(left int, right int) bool {
		return ids[left] < ids[right]
	})
	items := make([]worldfurniture.Item, 0, len(ids))
	for _, id := range ids {
		items = append(items, room.world.furniture[id])
	}

	return items
}

// SurfaceHeights returns current per-tile walkable heights in row-major order.
func (room *Room) SurfaceHeights() (uint16, uint16, []TileHeight) {
	room.mutex.RLock()
	defer room.mutex.RUnlock()
	if room.world == nil {
		return 0, 0, nil
	}
	width, height := room.world.grid.Width(), room.world.grid.Height()
	tiles := make([]TileHeight, 0, int(width)*int(height))
	for y := uint16(0); y < height; y++ {
		for x := uint16(0); x < width; x++ {
			section, err := room.world.resolver.TopSection(grid.Point{X: x, Y: y})
			if err != nil {
				tiles = append(tiles, TileHeight{})
				continue
			}
			tiles = append(tiles, TileHeight{Valid: true, Height: section.Z(), StackingBlocked: !section.Stacking()})
		}
	}

	return width, height, tiles
}

// SlotOccupant returns a player occupying a sit or lay slot of an item.
func (room *Room) SlotOccupant(itemID int64) (int64, bool) {
	room.mutex.RLock()
	defer room.mutex.RUnlock()
	if room.world == nil {
		return 0, false
	}
	item, found := room.world.furniture[itemID]
	if !found {
		return 0, false
	}
	for _, slot := range worldfurniture.Slots(item) {
		if playerID, occupied := room.world.slotOccupants[slot.Point]; occupied {
			return playerID, true
		}
	}

	return 0, false
}

// unitSnapshot maps a world unit to a runtime snapshot.
func unitSnapshot(playerID int64, roomUnit *worldunit.Unit) UnitSnapshot {
	return UnitSnapshot{
		PlayerID: playerID, UnitID: roomUnit.ID(), Position: roomUnit.Position(), Previous: roomUnit.Previous(),
		BodyRotation: roomUnit.BodyRotation(), HeadRotation: roomUnit.HeadRotation(),
		Moving: roomUnit.Moving(), Statuses: roomUnit.Statuses(),
	}
}
