// Package runtime owns mutable world state for one active room.
package runtime

import (
	"errors"
	"time"

	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	"github.com/niflaot/pixels/internal/realm/room/world/surface"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

var (
	// ErrWorldNotLoaded reports that room world behavior is unavailable.
	ErrWorldNotLoaded = errors.New("room world is not loaded")
	// ErrInvalidWorld reports invalid loaded world input.
	ErrInvalidWorld = errors.New("invalid room world")
	// ErrUnitNotFound reports a missing room world unit.
	ErrUnitNotFound = errors.New("room world unit not found")
	// ErrUnitExiting reports movement requested for a unit already leaving.
	ErrUnitExiting = errors.New("room world unit is exiting")
	// ErrInvalidPlacement reports a furniture footprint outside the room grid.
	ErrInvalidPlacement = errors.New("invalid furniture placement")
	// ErrTileOccupied reports a furniture footprint occupied by a room unit.
	ErrTileOccupied = errors.New("furniture placement tile is occupied")
	// ErrCannotStack reports a furniture footprint with a blocked stacking surface.
	ErrCannotStack = errors.New("furniture cannot stack on target surface")
	// ErrFurnitureNotFound reports a missing runtime furniture item.
	ErrFurnitureNotFound = errors.New("room world furniture not found")
)

// Config stores loaded room world input.
type Config struct {
	// Grid stores the immutable base room grid.
	Grid grid.Grid
	// Fixtures stores dynamic initial column fixtures.
	Fixtures []surface.Fixture
	// Furniture stores placed furniture items projected into fixtures on load.
	Furniture []worldfurniture.Item
	// Door stores the room entry position.
	Door worldpath.Position
	// Body stores the initial body rotation.
	Body worldunit.Rotation
	// Head stores the initial head rotation.
	Head worldunit.Rotation
	// Rules stores movement pathfinding rules.
	Rules worldpath.Rules
}

// UnitSnapshot stores stable world unit state.
type UnitSnapshot struct {
	// EntityKey stores the world index used by players, bots, and future pets.
	EntityKey int64
	// PlayerID stores the owning player id.
	PlayerID int64
	// OwnerID stores the durable owner for non-player units.
	OwnerID int64
	// Kind identifies the room unit family.
	Kind worldunit.Kind
	// UnitID stores the room-local unit id.
	UnitID int64
	// Position stores the current unit position.
	Position worldpath.Position
	// Previous stores the previous unit position.
	Previous worldpath.Position
	// BodyRotation stores the unit body rotation.
	BodyRotation worldunit.Rotation
	// HeadRotation stores the unit head rotation.
	HeadRotation worldunit.Rotation
	// Moving reports whether the unit is moving or awaiting its final settled projection.
	Moving bool
	// Statuses stores ordered unit statuses.
	Statuses []worldunit.Status
	// HandItem stores the currently carried protocol hand item id.
	HandItem int32
	// Idle reports whether the unit is projected as AFK.
	Idle bool
	// IdleSince stores when the current AFK projection began.
	IdleSince time.Time
	// ManualIdle reports whether explicit avatar activity must clear the AFK projection.
	ManualIdle bool
	// ActiveEffectID stores the selected avatar effect.
	ActiveEffectID int32
	// RenderOffset stores a protocol-only vertical offset from the physical tile.
	RenderOffset grid.Height
}

// Movement stores one unit tick movement.
type Movement struct {
	// PlayerID stores the owning player id.
	PlayerID int64
	// Unit stores the moved unit snapshot.
	Unit UnitSnapshot
	// Step stores the accepted path step.
	Step worldpath.Step
	// Moved reports whether the tick advanced to Step.
	Moved bool
	// Settled reports whether the tick cleared movement status.
	Settled bool
	// Exited reports whether the unit completed or could not continue a room exit.
	Exited bool
	// ForcedExit reports whether server moderation initiated the room exit.
	ForcedExit bool
}

// TileHeight describes one resolved tile's current walkable height and stacking state.
type TileHeight struct {
	// Valid reports whether the tile is part of the room.
	Valid bool
	// Height stores the current walkable top height.
	Height grid.Height
	// StackingBlocked reports whether new items cannot stack on this tile.
	StackingBlocked bool
}

// MovementPlan stores immutable pathfinding input captured from a world.
type MovementPlan struct {
	// Start stores the unit's current position.
	Start worldpath.Position
	// Goal stores the resolved target position.
	Goal grid.Point
	// Occupancy stores positions reserved by other units.
	Occupancy worldpath.Occupancy
}

// TeleportPolicy selects direct legacy placement or safe target-neighbor resolution.
type TeleportPolicy uint8

const (
	// TeleportDirect preserves an exact authoritative relocation.
	TeleportDirect TeleportPolicy = iota
	// TeleportNear selects the target or its first stable walkable unoccupied neighbor.
	TeleportNear
)

// teleportNear resolves a safe destination around one configured target.
func (world *World) teleportNear(playerID int64, roomUnit *worldunit.Unit, point grid.Point, rotation worldunit.Rotation, controlled bool) (UnitSnapshot, error) {
	candidates := [9]grid.Point{point}
	count := 1
	for direction := uint8(0); direction < 8; direction++ {
		if candidate, valid := grid.PointInFront(point, direction); valid {
			candidates[count], count = candidate, count+1
		}
	}
	for _, candidate := range candidates[:count] {
		section, err := world.resolver.TopSection(candidate)
		if err == nil && world.rules.AllowsSection(section) && !world.pointOccupied(playerID, candidate) {
			return world.repositionUnit(playerID, roomUnit, candidate, section.Z(), rotation, controlled), nil
		}
	}
	return UnitSnapshot{}, worldpath.ErrInvalidGoal
}

// repositionUnit commits one already-validated direct unit relocation.
func (world *World) repositionUnit(playerID int64, roomUnit *worldunit.Unit, point grid.Point, height grid.Height, rotation worldunit.Rotation, controlled bool) UnitSnapshot {
	world.releaseSlot(playerID)
	position := worldpath.Position{Point: point, Z: height}
	roomUnit.Reposition(position, rotation)
	if controlled {
		roomUnit.SetControl(worldunit.ControlTeleporting)
	} else {
		roomUnit.SetControl(worldunit.ControlNone)
	}
	if linkedKey, linked := world.linkedKey(playerID); linked {
		if linkedUnit, found := world.units[linkedKey]; found {
			world.releaseSlot(linkedKey)
			linkedUnit.Reposition(position, rotation)
			if controlled {
				linkedUnit.SetControl(worldunit.ControlTeleporting)
			} else {
				linkedUnit.SetControl(worldunit.ControlNone)
			}
		}
	}
	return unitSnapshot(playerID, roomUnit)
}

// UnitMotion returns allocation-free movement state for room-owned domain cycles.
func (world *World) UnitMotion(entityKey int64) (UnitSnapshot, bool) {
	roomUnit, found := world.units[entityKey]
	if !found {
		return UnitSnapshot{}, false
	}
	durablePlayerID := entityKey
	if roomUnit.Kind() != worldunit.KindPlayer {
		durablePlayerID = 0
	}
	return UnitSnapshot{
		EntityKey: entityKey, PlayerID: durablePlayerID, OwnerID: roomUnit.OwnerID(), Kind: roomUnit.Kind(), UnitID: roomUnit.ID(),
		Position: roomUnit.Position(), Previous: roomUnit.Previous(), BodyRotation: roomUnit.BodyRotation(), HeadRotation: roomUnit.HeadRotation(),
		Moving: roomUnit.InMotion(), HandItem: roomUnit.HandItem(), ActiveEffectID: roomUnit.ActiveEffect(), RenderOffset: roomUnit.RenderOffset(),
	}, true
}

// RandomWalkablePoint selects one walkable unoccupied tile with reservoir sampling and no allocation.
func (world *World) RandomWalkablePoint(entityKey int64, radius int, random uint64) (grid.Point, bool) {
	roomUnit, found := world.units[entityKey]
	if !found || radius <= 0 {
		return grid.Point{}, false
	}
	origin := roomUnit.Position().Point
	minX, maxX := boundedRange(int(origin.X), radius, int(world.grid.Width()))
	minY, maxY := boundedRange(int(origin.Y), radius, int(world.grid.Height()))
	selected := grid.Point{}
	count := uint64(0)
	for y := minY; y < maxY; y++ {
		for x := minX; x < maxX; x++ {
			point := grid.Point{X: uint16(x), Y: uint16(y)}
			if point == origin || world.pointOccupied(entityKey, point) {
				continue
			}
			section, err := world.resolver.TopSection(point)
			if err != nil || !world.rules.AllowsSection(section) {
				continue
			}
			count++
			if random%count == 0 {
				selected = point
			}
			random = random*6364136223846793005 + 1442695040888963407
		}
	}
	return selected, count > 0
}

// pointOccupied reports whether another unit currently owns a tile.
func (world *World) pointOccupied(excludedKey int64, point grid.Point) bool {
	linkedKey, linked := world.linkedKey(excludedKey)
	for key, roomUnit := range world.units {
		if key != excludedKey && (!linked || key != linkedKey) && roomUnit.Position().Point == point {
			return true
		}
	}
	return false
}

// boundedRange returns a half-open coordinate range.
func boundedRange(origin int, radius int, limit int) (int, int) {
	minimum := origin - radius
	if minimum < 0 {
		minimum = 0
	}
	maximum := origin + radius + 1
	if maximum > limit {
		maximum = limit
	}
	return minimum, maximum
}
