package live

import (
	"context"
	"time"

	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	"github.com/niflaot/pixels/internal/realm/room/world/surface"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	netconn "github.com/niflaot/pixels/networking/connection"
)

const (
	// DefaultTickInterval stores the default room movement tick interval.
	DefaultTickInterval = 500 * time.Millisecond
)

// Snapshot stores room metadata needed by runtime occupancy.
type Snapshot struct {
	// ID identifies the room.
	ID int64

	// OwnerPlayerID identifies the player with furniture management rights.
	OwnerPlayerID int64

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

	// Motto stores the visible player motto.
	Motto string

	// Figure stores the visible player figure.
	Figure string

	// Gender stores the visible player gender.
	Gender string

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

	// Moved reports whether the tick advanced to Step.
	Moved bool

	// Settled reports whether the tick cleared movement status.
	Settled bool
}

// MovementPublisher publishes room tick movements.
type MovementPublisher func(context.Context, *Room, []Movement) error

// RegistryOption configures a room registry.
type RegistryOption func(*Registry)

// WithMovementPublisher configures room movement publishing.
func WithMovementPublisher(publisher MovementPublisher) RegistryOption {
	return func(registry *Registry) {
		registry.movementPublish = publisher
	}
}

// WithTickInterval configures room movement tick cadence.
func WithTickInterval(interval time.Duration) RegistryOption {
	return func(registry *Registry) {
		if interval > 0 {
			registry.tickInterval = interval
		}
	}
}

// Valid reports whether the snapshot can back an active room.
func (snapshot Snapshot) Valid() bool {
	return snapshot.ID > 0 && snapshot.MaxUsers > 0
}

// CanManageFurniture reports whether a player may place, move, or pick up room furniture.
func (room *Room) CanManageFurniture(playerID int64) bool {
	room.mutex.RLock()
	defer room.mutex.RUnlock()

	return playerID > 0 && room.snapshot.OwnerPlayerID == playerID
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
