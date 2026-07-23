package runtime

import (
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldindex "github.com/niflaot/pixels/internal/realm/room/world/runtime/furnitureindex"
	worldmount "github.com/niflaot/pixels/internal/realm/room/world/runtime/mount"
	"github.com/niflaot/pixels/internal/realm/room/world/surface"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

var (
	// ErrInvalidMount reports incompatible or already-linked riding units.
	ErrInvalidMount = worldmount.ErrInvalid
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
	// mounts stores bidirectional rider and mounted-unit links.
	mounts *worldmount.State
}

// New creates loaded world state.
func New(config Config) (*World, error) {
	indexes, err := worldindex.Build(config.Furniture)
	if err != nil {
		return nil, err
	}
	fixtures := append(indexes.Fixtures, config.Fixtures...)
	resolver, err := surface.NewResolver(config.Grid, fixtures)
	if err != nil {
		return nil, err
	}
	if _, err := resolver.SectionAt(config.Door.Point, config.Door.Z); err != nil {
		return nil, ErrInvalidWorld
	}

	return &World{
		grid: config.Grid, resolver: resolver, furniture: indexes.Items, furnitureTiles: indexes.Tiles,
		interactionTypes: indexes.Types, interactions: indexes.Interactions,
		door: config.Door, body: config.Body, head: config.Head, rules: config.Rules.Normalize(),
		units: make(map[int64]*worldunit.Unit), unitKeys: make(map[int64]int64), nextUnitID: 1,
		unitSlots: make(map[int64]grid.Point), slotOccupants: make(map[grid.Point]int64),
		mounts: worldmount.New(),
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
	world.unlinkMount(entityKey)
	world.releaseSlot(entityKey)
	delete(world.unitKeys, roomUnit.ID())
	delete(world.units, entityKey)
	return snapshot, true
}

// RemoveUnit removes a player world unit and releases its furniture slot.
func (world *World) RemoveUnit(playerID int64) {
	world.unlinkMount(playerID)
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
	world.mounts = worldmount.New()
}

// SetMount attaches or detaches one player unit and one pet unit.
func (world *World) SetMount(riderKey int64, mountKey int64, mounted bool) (UnitSnapshot, UnitSnapshot, error) {
	rider, mountUnit, err := world.mounts.Set(world.units, riderKey, mountKey, mounted)
	if err != nil {
		return UnitSnapshot{}, UnitSnapshot{}, ErrInvalidMount
	}
	if rider == nil || mountUnit == nil {
		return UnitSnapshot{}, UnitSnapshot{}, ErrUnitNotFound
	}
	world.releaseSlot(riderKey)
	world.releaseSlot(mountKey)
	return unitSnapshot(riderKey, rider), unitSnapshot(mountKey, mountUnit), nil
}

// unlinkMount clears either side of one rider and pet movement relationship.
func (world *World) unlinkMount(entityKey int64) {
	world.mounts.Unlink(world.units, entityKey)
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
