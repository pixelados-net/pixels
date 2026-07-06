package live

import (
	"sort"

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

// MoveTo sets a unit movement goal.
func (room *Room) MoveTo(playerID int64, goal grid.Point) (worldpath.Path, error) {
	runtime, start, occupancy, err := room.movementSnapshot(playerID)
	if err != nil {
		return worldpath.Path{}, err
	}

	finder := worldpath.NewFinderWithOccupancy(runtime.resolver, runtime.rules, occupancy)
	roomPath, err := finder.Find(start, goal)
	if err != nil {
		return worldpath.Path{}, err
	}

	room.mutex.Lock()
	defer room.mutex.Unlock()

	if room.world != runtime {
		return worldpath.Path{}, worldpath.ErrInvalidPath
	}
	roomUnit, ok := room.world.units[playerID]
	if !ok {
		return worldpath.Path{}, ErrUnitNotFound
	}
	if err := roomPath.Validate(room.world.resolver); err != nil {
		return worldpath.Path{}, err
	}
	roomUnit.SetPath(roomPath)

	return roomPath, nil
}

// Tick advances room world movement once.
func (room *Room) Tick() []Movement {
	room.mutex.Lock()
	defer room.mutex.Unlock()

	if room.world == nil {
		return nil
	}

	playerIDs := room.world.sortedPlayerIDs()
	movements := make([]Movement, 0, len(playerIDs))
	for _, playerID := range playerIDs {
		roomUnit := room.world.units[playerID]
		step, moved := roomUnit.Advance()
		if !moved {
			continue
		}
		movements = append(movements, Movement{PlayerID: playerID, Unit: unitSnapshot(playerID, roomUnit), Step: step})
	}

	return movements
}

// movementSnapshot returns data needed to calculate movement outside the room lock.
func (room *Room) movementSnapshot(playerID int64) (*World, worldpath.Position, worldpath.Occupancy, error) {
	room.mutex.RLock()
	defer room.mutex.RUnlock()

	if room.world == nil {
		return nil, worldpath.Position{}, worldpath.Occupancy{}, ErrWorldNotLoaded
	}
	roomUnit, ok := room.world.units[playerID]
	if !ok {
		return nil, worldpath.Position{}, worldpath.Occupancy{}, ErrUnitNotFound
	}

	return room.world, roomUnit.Position(), room.world.occupancyExcept(playerID), nil
}

// newWorld creates loaded world state.
func newWorld(config WorldConfig) (*World, error) {
	resolver, err := surface.NewResolver(config.Grid, config.Fixtures)
	if err != nil {
		return nil, err
	}
	if _, err := resolver.SectionAt(config.Door.Point, config.Door.Z); err != nil {
		return nil, ErrInvalidWorld
	}

	return &World{
		grid:       config.Grid,
		resolver:   resolver,
		door:       config.Door,
		body:       config.Body,
		head:       config.Head,
		rules:      config.Rules.Normalize(),
		units:      make(map[int64]*worldunit.Unit),
		nextUnitID: 1,
	}, nil
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
	delete(world.units, playerID)
}

// clearUnits removes all world units.
func (world *World) clearUnits() {
	world.units = make(map[int64]*worldunit.Unit)
}

// occupancyExcept returns occupied positions except one player.
func (world *World) occupancyExcept(playerID int64) worldpath.Occupancy {
	positions := make([]worldpath.Position, 0, len(world.units))
	for occupantID, roomUnit := range world.units {
		if occupantID == playerID {
			continue
		}
		positions = append(positions, roomUnit.Position())
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
