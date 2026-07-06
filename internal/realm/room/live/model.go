package live

import (
	"time"

	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	"github.com/niflaot/pixels/internal/realm/room/world/surface"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	netconn "github.com/niflaot/pixels/networking/connection"
)

// Snapshot stores room metadata needed by runtime occupancy.
type Snapshot struct {
	// ID identifies the room.
	ID int64

	// CategoryID optionally identifies the room category.
	CategoryID *int64

	// MaxUsers stores the maximum active occupancy.
	MaxUsers int
}

// Occupancy describes active room population.
type Occupancy struct {
	// RoomID identifies the room.
	RoomID int64

	// CategoryID optionally identifies the room category.
	CategoryID *int64

	// Count stores the active occupancy count.
	Count int

	// MaxUsers stores the maximum active occupancy.
	MaxUsers int

	// PlayerIDs stores active player ids.
	PlayerIDs []int64
}

// Occupant describes one player inside an active room.
type Occupant struct {
	// PlayerID identifies the player.
	PlayerID int64

	// Username stores a display snapshot for diagnostics.
	Username string

	// ConnectionID identifies the active connection.
	ConnectionID netconn.ID

	// ConnectionKind identifies the active connection family.
	ConnectionKind netconn.Kind

	// JoinedAt stores when the player joined the active room.
	JoinedAt time.Time
}

// WorldConfig stores loaded room world input.
type WorldConfig struct {
	// Grid stores the immutable base room grid.
	Grid grid.Grid

	// Fixtures stores dynamic initial column fixtures.
	Fixtures []surface.Fixture

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

	// Moving reports whether the unit has pending steps.
	Moving bool

	// Statuses stores ordered unit statuses.
	Statuses []worldunit.Status
}

// Movement stores one unit tick movement.
type Movement struct {
	// PlayerID stores the owning player id.
	PlayerID int64

	// Unit stores the moved unit snapshot.
	Unit UnitSnapshot

	// Step stores the accepted path step.
	Step worldpath.Step
}

// Valid reports whether the snapshot can back an active room.
func (snapshot Snapshot) Valid() bool {
	return snapshot.ID > 0 && snapshot.MaxUsers > 0
}

// Valid reports whether the occupant can join a room.
func (occupant Occupant) Valid() bool {
	return occupant.PlayerID > 0 && occupant.ConnectionID != "" && occupant.ConnectionKind != ""
}

// WithJoinTime returns the occupant with a default join time.
func (occupant Occupant) WithJoinTime(now time.Time) Occupant {
	if occupant.JoinedAt.IsZero() {
		occupant.JoinedAt = now
	}

	return occupant
}
