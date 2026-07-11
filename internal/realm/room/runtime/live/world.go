package live

import (
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldruntime "github.com/niflaot/pixels/internal/realm/room/world/runtime"
	"github.com/niflaot/pixels/internal/realm/room/world/surface"
)

// LoadWorld loads or replaces room world behavior.
func (room *Room) LoadWorld(config WorldConfig) error {
	loaded, err := worldruntime.New(config)
	if err != nil {
		return err
	}
	room.mutex.Lock()
	defer room.mutex.Unlock()
	room.world = loaded
	for playerID := range room.occupants {
		room.world.AddUnit(playerID)
	}

	return nil
}

// ReloadFixtures replaces one source's fixtures without resetting units.
func (room *Room) ReloadFixtures(sourceID int64, fixtures []surface.Fixture) error {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	if room.world == nil {
		return ErrWorldNotLoaded
	}

	return room.world.ReloadFixtures(sourceID, fixtures)
}

// ReloadFurniture replaces one furniture item and returns affected units.
func (room *Room) ReloadFurniture(sourceID int64, item *worldfurniture.Item) ([]UnitSnapshot, error) {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	if room.world == nil {
		return nil, ErrWorldNotLoaded
	}

	return room.world.ReloadFurniture(sourceID, item)
}

// UnloadWorld unloads room world behavior.
func (room *Room) UnloadWorld() {
	room.mutex.Lock()
	room.world = nil
	room.mutex.Unlock()
}

// WorldLoaded reports whether world behavior is loaded.
func (room *Room) WorldLoaded() bool {
	room.mutex.RLock()
	loaded := room.world != nil
	room.mutex.RUnlock()

	return loaded
}

// Units returns stable world unit snapshots.
func (room *Room) Units() []UnitSnapshot {
	room.mutex.RLock()
	defer room.mutex.RUnlock()
	if room.world == nil {
		return nil
	}

	return room.world.Units()
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
		unit, found := room.world.Unit(playerID)
		if found {
			presences = append(presences, Presence{Occupant: occupant, Unit: unit})
		}
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

	return room.world.Unit(playerID)
}

// SetUnitStatus stores one status on a player unit when its world is loaded.
func (room *Room) SetUnitStatus(playerID int64, key string, value string) bool {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	if room.world == nil {
		return false
	}

	return room.world.SetUnitStatus(playerID, key, value)
}

// FurnitureItems returns stable placed furniture item snapshots.
func (room *Room) FurnitureItems() []worldfurniture.Item {
	room.mutex.RLock()
	defer room.mutex.RUnlock()
	if room.world == nil {
		return nil
	}

	return room.world.FurnitureItems()
}

// SurfaceHeights returns current per-tile walkable heights in row-major order.
func (room *Room) SurfaceHeights() (uint16, uint16, []TileHeight) {
	room.mutex.RLock()
	defer room.mutex.RUnlock()
	if room.world == nil {
		return 0, 0, nil
	}

	return room.world.SurfaceHeights()
}

// SlotOccupant returns a player occupying a sit or lay slot of an item.
func (room *Room) SlotOccupant(itemID int64) (int64, bool) {
	room.mutex.RLock()
	defer room.mutex.RUnlock()
	if room.world == nil {
		return 0, false
	}

	return room.world.SlotOccupant(itemID)
}

// ResolveFurniturePlacement validates a footprint against occupancy and stacking rules.
func (room *Room) ResolveFurniturePlacement(sourceID int64, footprint []grid.Point) (grid.Height, error) {
	room.mutex.RLock()
	defer room.mutex.RUnlock()
	if room.world == nil {
		return 0, ErrWorldNotLoaded
	}

	return room.world.ResolveFurniturePlacement(sourceID, footprint)
}
