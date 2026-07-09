package live

import (
	"fmt"
	"sort"

	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	"github.com/niflaot/pixels/internal/realm/room/world/surface"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// World stores loaded room world state.
type World struct {
	// grid stores immutable world grid.
	grid grid.Grid

	// resolver resolves world columns.
	resolver *surface.Resolver

	// furniture stores placed furniture items by id.
	furniture map[int64]worldfurniture.Item

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

	// nextUnitID stores the next room-local unit id.
	nextUnitID int64

	// unitSlots stores the slot tile a player currently occupies, keyed by player id.
	unitSlots map[int64]grid.Point

	// slotOccupants stores the player id occupying a slot, keyed by slot tile.
	slotOccupants map[grid.Point]int64
}

// LoadWorld loads or replaces room world behavior.
func (room *Room) LoadWorld(config WorldConfig) error {
	loaded, err := newWorld(config)
	if err != nil {
		return err
	}

	room.mutex.Lock()
	defer room.mutex.Unlock()

	room.world = loaded
	for playerID := range room.occupants {
		room.world.addUnit(playerID)
	}

	return nil
}

// ReloadFixtures replaces the room world fixture set for one source without resetting existing units.
func (room *Room) ReloadFixtures(sourceID int64, fixtures []surface.Fixture) error {
	room.mutex.Lock()
	defer room.mutex.Unlock()

	if room.world == nil {
		return ErrWorldNotLoaded
	}

	return room.world.resolver.ReplaceFixtures(sourceID, fixtures)
}

// ReloadFurniture replaces the world fixtures and tracked snapshot for one furniture item, or removes
// it entirely when item is nil, without resetting existing units. Any sit/lay occupants of the
// item's previous slots are reconciled against the updated item: an occupant whose slot tile still
// exists (e.g. a rotation that keeps the same anchor tile) is re-settled facing the new direction,
// while an occupant whose slot tile disappeared (the item moved away or was picked up) stands up.
// The returned snapshots are every unit whose status changed, for the caller to broadcast.
func (room *Room) ReloadFurniture(sourceID int64, item *worldfurniture.Item) ([]UnitSnapshot, error) {
	room.mutex.Lock()
	defer room.mutex.Unlock()

	if room.world == nil {
		return nil, ErrWorldNotLoaded
	}

	var fixtures []surface.Fixture
	if item != nil {
		built, err := worldfurniture.Fixtures(*item)
		if err != nil {
			return nil, err
		}
		fixtures = built
	}

	previous, hadPrevious := room.world.furniture[sourceID]
	if err := room.world.resolver.ReplaceFixtures(sourceID, fixtures); err != nil {
		return nil, err
	}

	if item != nil {
		room.world.furniture[sourceID] = *item
	} else {
		delete(room.world.furniture, sourceID)
	}

	if !hadPrevious {
		return nil, nil
	}

	return room.world.reconcileSlotOccupants(worldfurniture.Slots(previous), item), nil
}

// UnloadWorld unloads room world behavior.
func (room *Room) UnloadWorld() {
	room.mutex.Lock()
	defer room.mutex.Unlock()

	room.world = nil
}

// WorldLoaded reports whether world behavior is loaded.
func (room *Room) WorldLoaded() bool {
	room.mutex.RLock()
	defer room.mutex.RUnlock()

	return room.world != nil
}

// newWorld creates loaded world state.
func newWorld(config WorldConfig) (*World, error) {
	fixtures, furnitureIndex, err := furnitureFixtures(config.Furniture)
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
		grid:          config.Grid,
		resolver:      resolver,
		furniture:     furnitureIndex,
		door:          config.Door,
		body:          config.Body,
		head:          config.Head,
		rules:         config.Rules.Normalize(),
		units:         make(map[int64]*worldunit.Unit),
		nextUnitID:    1,
		unitSlots:     make(map[int64]grid.Point),
		slotOccupants: make(map[grid.Point]int64),
	}, nil
}

// furnitureFixtures converts placed furniture items into resolver fixtures and an id index.
func furnitureFixtures(items []worldfurniture.Item) ([]surface.Fixture, map[int64]worldfurniture.Item, error) {
	fixtures := make([]surface.Fixture, 0, len(items))
	index := make(map[int64]worldfurniture.Item, len(items))
	for _, item := range items {
		itemFixtures, err := worldfurniture.Fixtures(item)
		if err != nil {
			return nil, nil, fmt.Errorf("build fixtures for furniture item %d: %w", item.ID, err)
		}
		fixtures = append(fixtures, itemFixtures...)
		index[item.ID] = item
	}

	return fixtures, index, nil
}

// addUnit creates a world unit for a player.
func (world *World) addUnit(playerID int64) {
	if _, exists := world.units[playerID]; exists {
		return
	}

	roomUnit, err := worldunit.New(worldunit.Params{
		ID:       world.nextUnitID,
		OwnerID:  playerID,
		Kind:     worldunit.KindPlayer,
		Position: world.door,
		Body:     world.body,
		Head:     world.head,
	})
	if err != nil {
		return
	}
	world.units[playerID] = roomUnit
	world.nextUnitID++
}

// removeUnit removes a player world unit.
func (world *World) removeUnit(playerID int64) {
	world.releaseSlot(playerID)
	delete(world.units, playerID)
}

// clearUnits removes all world units.
func (world *World) clearUnits() {
	world.units = make(map[int64]*worldunit.Unit)
	world.unitSlots = make(map[int64]grid.Point)
	world.slotOccupants = make(map[grid.Point]int64)
}

// resolveSlotGoal snaps a goal inside a slotted furniture item's footprint onto the item's anchor
// slot, matching the goal's column or row so multi-seat items keep independent anchors. Goals that
// already target a slot tile, or items without slots, are left untouched.
func (world *World) resolveSlotGoal(goal grid.Point) grid.Point {
	for _, item := range world.furniture {
		slots := worldfurniture.Slots(item)
		if len(slots) == 0 || !footprintContains(item, goal) {
			continue
		}
		for _, slot := range slots {
			if slot.Point == goal {
				return goal
			}
		}
		for _, slot := range slots {
			if slot.Point.X == goal.X || slot.Point.Y == goal.Y {
				return slot.Point
			}
		}

		return slots[0].Point
	}

	return goal
}

// footprintContains reports whether a point falls inside a placed item's rotated footprint.
func footprintContains(item worldfurniture.Item, point grid.Point) bool {
	for _, tile := range worldfurniture.Footprint(item.Point, item.Definition.Width, item.Definition.Length, item.Rotation) {
		if tile == point {
			return true
		}
	}

	return false
}

// slotAt finds a placed furniture item's slot at a tile with a matching status.
func (world *World) slotAt(point grid.Point, status worldfurniture.SlotStatus) (worldfurniture.Slot, bool) {
	for _, item := range world.furniture {
		for _, slot := range worldfurniture.Slots(item) {
			if slot.Point == point && slot.Status == status {
				return slot, true
			}
		}
	}

	return worldfurniture.Slot{}, false
}

// occupySlot records that a player now occupies a slot tile, replacing any prior slot.
func (world *World) occupySlot(playerID int64, point grid.Point) {
	world.releaseSlot(playerID)
	world.unitSlots[playerID] = point
	world.slotOccupants[point] = playerID
}

// releaseSlot clears a player's occupied slot tile, if any.
func (world *World) releaseSlot(playerID int64) {
	point, ok := world.unitSlots[playerID]
	if !ok {
		return
	}
	delete(world.unitSlots, playerID)
	delete(world.slotOccupants, point)
}

// reconcileSlotOccupants re-settles or stands up units occupying a furniture item's previous slots
// after the item changes. A unit whose slot point still resolves to a slot on the updated item
// (item is non-nil and declares a slot at that exact point) is re-settled with the new slot's
// rotation and height offset; otherwise it stands up and its slot is released.
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

// slotAtPoint finds a slot at an exact point.
func slotAtPoint(slots []worldfurniture.Slot, point grid.Point) (worldfurniture.Slot, bool) {
	for _, slot := range slots {
		if slot.Point == point {
			return slot, true
		}
	}

	return worldfurniture.Slot{}, false
}

// unitStatusFor maps a slot status to its unit status key.
func unitStatusFor(status worldfurniture.SlotStatus) string {
	if status == worldfurniture.SlotStatusLay {
		return worldunit.StatusLay
	}

	return worldunit.StatusSit
}

// occupancyExcept returns occupied positions except one player.
func (world *World) occupancyExcept(playerID int64) worldpath.Occupancy {
	positions := make([]worldpath.Position, 0, len(world.units)*2)
	for occupantID, roomUnit := range world.units {
		if occupantID == playerID {
			continue
		}
		positions = append(positions, roomUnit.Position())
		if goal, ok := roomUnit.Goal(); ok {
			positions = append(positions, goal)
		}
	}

	return worldpath.NewOccupancy(positions)
}

// sortedPlayerIDs returns world player ids in stable order.
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
