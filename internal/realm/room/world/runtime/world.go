package runtime

import (
	"fmt"

	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	"github.com/niflaot/pixels/internal/realm/room/world/surface"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// World stores loaded room world state.
type World struct {
	// grid stores the immutable world grid.
	grid grid.Grid
	// resolver resolves world columns.
	resolver *surface.Resolver
	// furniture stores placed furniture items by id.
	furniture map[int64]worldfurniture.Item
	// furnitureTiles indexes furniture footprints by tile.
	furnitureTiles map[grid.Point][]int64
	// interactionTypes indexes furniture ids by interaction type.
	interactionTypes map[string][]int64
	// interactions indexes interactive furniture by footprint tile.
	interactions map[grid.Point]int64
	// door stores the room entry position.
	door worldpath.Position
	// body stores the initial body rotation.
	body worldunit.Rotation
	// head stores the initial head rotation.
	head worldunit.Rotation
	// rules stores movement pathfinding rules.
	rules worldpath.Rules
	// units stores world units by player id.
	units map[int64]*worldunit.Unit
	// unitKeys resolves room-local unit ids to entity keys without scanning occupants.
	unitKeys map[int64]int64
	// nextUnitID stores the next room-local unit id.
	nextUnitID int64
	// unitSlots stores the slot tile occupied by each player.
	unitSlots map[int64]grid.Point
	// slotOccupants stores the player occupying each slot tile.
	slotOccupants map[grid.Point]int64
}

// New creates loaded world state.
func New(config Config) (*World, error) {
	fixtures, furnitureIndex, tileIndex, typeIndex, interactionIndex, err := furnitureFixtures(config.Furniture)
	if err != nil {
		return nil, err
	}
	fixtures = append(fixtures, config.Fixtures...)
	resolver, err := surface.NewResolver(config.Grid, fixtures)
	if err != nil {
		return nil, err
	}
	if _, err := resolver.SectionAt(config.Door.Point, config.Door.Z); err != nil {
		return nil, ErrInvalidWorld
	}

	return &World{
		grid: config.Grid, resolver: resolver, furniture: furnitureIndex, furnitureTiles: tileIndex,
		interactionTypes: typeIndex, interactions: interactionIndex,
		door: config.Door, body: config.Body, head: config.Head, rules: config.Rules.Normalize(),
		units: make(map[int64]*worldunit.Unit), unitKeys: make(map[int64]int64), nextUnitID: 1,
		unitSlots: make(map[int64]grid.Point), slotOccupants: make(map[grid.Point]int64),
	}, nil
}

// AddUnit creates a world unit for a player when absent.
func (world *World) AddUnit(playerID int64) {
	if _, exists := world.units[playerID]; exists {
		return
	}
	roomUnit, err := worldunit.New(worldunit.Params{
		ID: world.nextUnitID, OwnerID: playerID, Kind: worldunit.KindPlayer,
		Position: world.door, Body: world.body, Head: world.head,
	})
	if err != nil {
		return
	}
	world.units[playerID] = roomUnit
	world.unitKeys[roomUnit.ID()] = playerID
	world.nextUnitID++
}

// AddEntity creates a non-player unit at an explicit resolved position.
func (world *World) AddEntity(entityKey int64, ownerID int64, kind worldunit.Kind, position worldpath.Position, rotation worldunit.Rotation) (UnitSnapshot, error) {
	if entityKey == 0 || ownerID <= 0 || kind == worldunit.KindPlayer {
		return UnitSnapshot{}, ErrInvalidWorld
	}
	if _, exists := world.units[entityKey]; exists {
		return unitSnapshot(entityKey, world.units[entityKey]), nil
	}
	section, err := world.resolver.TopSection(position.Point)
	if err != nil || !world.rules.AllowsSection(section) {
		return UnitSnapshot{}, worldpath.ErrInvalidGoal
	}
	for _, current := range world.units {
		if current.Position().Point == position.Point {
			return UnitSnapshot{}, ErrTileOccupied
		}
	}
	position.Z = section.Z()
	roomUnit, err := worldunit.New(worldunit.Params{ID: world.nextUnitID, OwnerID: ownerID, Kind: kind, Position: position, Body: rotation, Head: rotation})
	if err != nil {
		return UnitSnapshot{}, err
	}
	world.units[entityKey] = roomUnit
	world.unitKeys[roomUnit.ID()] = entityKey
	world.nextUnitID++
	return unitSnapshot(entityKey, roomUnit), nil
}

// RemoveEntity removes a non-player unit and releases its reserved slot.
func (world *World) RemoveEntity(entityKey int64) (UnitSnapshot, bool) {
	roomUnit, found := world.units[entityKey]
	if !found {
		return UnitSnapshot{}, false
	}
	snapshot := unitSnapshot(entityKey, roomUnit)
	world.releaseSlot(entityKey)
	delete(world.unitKeys, roomUnit.ID())
	delete(world.units, entityKey)
	return snapshot, true
}

// RemoveUnit removes a player world unit and releases its furniture slot.
func (world *World) RemoveUnit(playerID int64) {
	world.releaseSlot(playerID)
	if roomUnit, found := world.units[playerID]; found {
		delete(world.unitKeys, roomUnit.ID())
	}
	delete(world.units, playerID)
}

// ClearUnits removes every world unit and slot reservation.
func (world *World) ClearUnits() {
	world.units = make(map[int64]*worldunit.Unit)
	world.unitKeys = make(map[int64]int64)
	world.unitSlots = make(map[int64]grid.Point)
	world.slotOccupants = make(map[grid.Point]int64)
}

// Door returns the room entry position.
func (world *World) Door() worldpath.Position {
	return world.door
}

// ReloadFixtures replaces fixtures for one source.
func (world *World) ReloadFixtures(sourceID int64, fixtures []surface.Fixture) error {
	return world.resolver.ReplaceFixtures(sourceID, fixtures)
}

// ReloadFurniture replaces one furniture snapshot and reconciles occupied slots.
func (world *World) ReloadFurniture(sourceID int64, item *worldfurniture.Item) ([]UnitSnapshot, error) {
	var fixtures []surface.Fixture
	if item != nil {
		built, err := worldfurniture.Fixtures(*item)
		if err != nil {
			return nil, err
		}
		fixtures = built
	}
	previous, hadPrevious := world.furniture[sourceID]
	if err := world.resolver.ReplaceFixtures(sourceID, fixtures); err != nil {
		return nil, err
	}
	if item != nil {
		world.furniture[sourceID] = *item
	} else {
		delete(world.furniture, sourceID)
	}
	if hadPrevious {
		world.removeInteraction(previous)
		world.removeFurnitureIndexes(previous)
	}
	if item != nil {
		world.addInteraction(*item)
		world.addFurnitureIndexes(*item)
	}
	if !hadPrevious {
		return nil, nil
	}

	return world.reconcileSlotOccupants(worldfurniture.Slots(previous), item), nil
}

// furnitureFixtures converts furniture into resolver fixtures and an id index.
func furnitureFixtures(items []worldfurniture.Item) ([]surface.Fixture, map[int64]worldfurniture.Item, map[grid.Point][]int64, map[string][]int64, map[grid.Point]int64, error) {
	fixtures := make([]surface.Fixture, 0, len(items))
	index := make(map[int64]worldfurniture.Item, len(items))
	tiles := make(map[grid.Point][]int64)
	types := make(map[string][]int64)
	var interactions map[grid.Point]int64
	for _, item := range items {
		itemFixtures, err := worldfurniture.Fixtures(item)
		if err != nil {
			return nil, nil, nil, nil, nil, fmt.Errorf("build fixtures for furniture item %d: %w", item.ID, err)
		}
		fixtures = append(fixtures, itemFixtures...)
		index[item.ID] = item
		for _, point := range worldfurniture.Footprint(item.Point, item.Definition.Width, item.Definition.Length, item.Rotation) {
			tiles[point] = append(tiles[point], item.ID)
		}
		if item.Definition.InteractionType != "" {
			types[item.Definition.InteractionType] = append(types[item.Definition.InteractionType], item.ID)
		}
		if interactive(item) {
			if interactions == nil {
				interactions = make(map[grid.Point]int64)
			}
			for _, point := range worldfurniture.Footprint(item.Point, item.Definition.Width, item.Definition.Length, item.Rotation) {
				interactions[point] = item.ID
			}
		}
	}

	return fixtures, index, tiles, types, interactions, nil
}

// addInteraction indexes one interactive furniture footprint.
func (world *World) addInteraction(item worldfurniture.Item) {
	if !interactive(item) {
		return
	}
	if world.interactions == nil {
		world.interactions = make(map[grid.Point]int64)
	}
	for _, point := range worldfurniture.Footprint(item.Point, item.Definition.Width, item.Definition.Length, item.Rotation) {
		world.interactions[point] = item.ID
	}
}

// removeInteraction removes one interactive furniture footprint.
func (world *World) removeInteraction(item worldfurniture.Item) {
	for _, point := range worldfurniture.Footprint(item.Point, item.Definition.Width, item.Definition.Length, item.Rotation) {
		if world.interactions[point] == item.ID {
			delete(world.interactions, point)
		}
	}
	if len(world.interactions) == 0 {
		world.interactions = nil
	}
}
