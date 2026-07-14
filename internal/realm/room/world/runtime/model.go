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
	// PlayerID stores the owning player id.
	PlayerID int64
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
