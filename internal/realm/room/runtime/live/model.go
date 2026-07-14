// Package live contains active room lifecycle and registry state.
package live

import (
	"context"
	"errors"
	"time"

	roomdoorbell "github.com/niflaot/pixels/internal/realm/room/access/doorbell"
	worldruntime "github.com/niflaot/pixels/internal/realm/room/world/runtime"
	netconn "github.com/niflaot/pixels/networking/connection"
)

const (
	// DefaultTickInterval stores the default room movement tick interval.
	DefaultTickInterval = 500 * time.Millisecond
)

var (
	// ErrInvalidRoom reports malformed active room data.
	ErrInvalidRoom = errors.New("invalid active room")
	// ErrRoomClosed reports a closed active room.
	ErrRoomClosed = errors.New("room closed")
	// ErrRoomNotFound reports a missing active room.
	ErrRoomNotFound = errors.New("active room not found")
	// ErrInvalidOccupant reports malformed occupant data.
	ErrInvalidOccupant = errors.New("invalid room occupant")
	// ErrRoomFull reports an active room at capacity.
	ErrRoomFull = errors.New("room full")
	// ErrWorldNotLoaded reports world behavior before room world loading.
	ErrWorldNotLoaded = worldruntime.ErrWorldNotLoaded
	// ErrInvalidWorld reports malformed room world loading input.
	ErrInvalidWorld = worldruntime.ErrInvalidWorld
	// ErrUnitNotFound reports a missing room world unit.
	ErrUnitNotFound = worldruntime.ErrUnitNotFound
	// ErrUnitExiting reports client movement while a server-controlled exit is active.
	ErrUnitExiting = worldruntime.ErrUnitExiting
	// ErrInvalidPlacement reports a footprint tile outside the room grid.
	ErrInvalidPlacement = worldruntime.ErrInvalidPlacement
	// ErrTileOccupied reports a footprint tile currently occupied by a unit.
	ErrTileOccupied = worldruntime.ErrTileOccupied
	// ErrCannotStack reports a footprint tile that does not accept stacking.
	ErrCannotStack = worldruntime.ErrCannotStack
	// ErrNoFurnitureRights reports furniture management without room rights.
	ErrNoFurnitureRights = errors.New("player has no furniture rights in room")
)

// WorldConfig aliases mutable world loading input.
type WorldConfig = worldruntime.Config

// UnitSnapshot aliases stable world unit state.
type UnitSnapshot = worldruntime.UnitSnapshot

// Movement aliases one unit tick movement.
type Movement = worldruntime.Movement

// TileHeight aliases one resolved tile height.
type TileHeight = worldruntime.TileHeight

// Snapshot stores room metadata needed by runtime occupancy.
type Snapshot struct {
	// ID identifies the room.
	ID int64
	// OwnerPlayerID identifies the room owner.
	OwnerPlayerID int64
	// CategoryID optionally identifies the room category.
	CategoryID *int64
	// MaxUsers stores the maximum active occupancy.
	MaxUsers int
	// TradeMode describes direct-trade behavior in the room.
	TradeMode int16
	// ChatDistance stores the normal chat hearing radius in tiles.
	ChatDistance int16
	// ChatProtection stores the room flood-control tier.
	ChatProtection int16
}

// Presence pairs one active occupant with its room unit state.
type Presence struct {
	// Occupant stores connection and visible identity state.
	Occupant Occupant
	// Unit stores position and room-local unit state.
	Unit UnitSnapshot
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
	// ActiveEffectID stores the selected avatar effect.
	ActiveEffectID int32
	// ConnectionID identifies the active connection.
	ConnectionID netconn.ID
	// ConnectionKind identifies the active connection family.
	ConnectionKind netconn.Kind
	// JoinedAt stores when the player joined the active room.
	JoinedAt time.Time
}

// MovementPublisher publishes room tick movements.
type MovementPublisher func(context.Context, *Room, []Movement) error

// DoorbellPublisher publishes expired room entry requests.
type DoorbellPublisher func(context.Context, *Room, []roomdoorbell.Expired) error

// DoorbellApprover reports whether an authorized responder remains in a room.
type DoorbellApprover func(context.Context, *Room) (bool, error)

// CyclePublisher advances optional room-owned domain cycles.
type CyclePublisher func(context.Context, *Room, time.Time) error

// SetCyclePublisher configures the optional domain cycle before rooms activate.
func (registry *Registry) SetCyclePublisher(publisher CyclePublisher) {
	registry.mutex.Lock()
	registry.cyclePublish = publisher
	registry.mutex.Unlock()
}

// RegistryOption configures a room registry.
type RegistryOption func(*Registry)

// WithMovementPublisher configures room movement publishing.
func WithMovementPublisher(publisher MovementPublisher) RegistryOption {
	return func(registry *Registry) { registry.movementPublish = publisher }
}

// WithDoorbellPublisher configures doorbell expiration publishing.
func WithDoorbellPublisher(publisher DoorbellPublisher) RegistryOption {
	return func(registry *Registry) { registry.doorbellPublish = publisher }
}

// WithDoorbellApprover configures authorized responder presence checks.
func WithDoorbellApprover(approver DoorbellApprover) RegistryOption {
	return func(registry *Registry) { registry.doorbellApprover = approver }
}

// WithDoorbellTimeout configures waiting request duration.
func WithDoorbellTimeout(timeout time.Duration) RegistryOption {
	return func(registry *Registry) {
		if timeout > 0 {
			registry.doorbellTimeout = timeout
		}
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

// UpdateSettings refreshes runtime metadata affected by persistent settings.
func (room *Room) UpdateSettings(categoryID *int64, maxUsers int, chatDistance int16, chatProtection int16) {
	room.mutex.Lock()
	room.snapshot.CategoryID = categoryID
	room.snapshot.MaxUsers = maxUsers
	room.snapshot.ChatDistance = chatDistance
	room.snapshot.ChatProtection = chatProtection
	room.mutex.Unlock()
}

// UpdateCategoryAndTrade refreshes focused category and trade policy metadata.
func (room *Room) UpdateCategoryAndTrade(categoryID *int64, tradeMode int16) {
	room.mutex.Lock()
	room.snapshot.CategoryID = categoryID
	room.snapshot.TradeMode = tradeMode
	room.mutex.Unlock()
}

// SetMuteAll replaces the active room mute-all state.
func (room *Room) SetMuteAll(muted bool) {
	room.muteAll.Store(muted)
}

// MuteAll reports whether non-privileged room chat is globally muted.
func (room *Room) MuteAll() bool {
	return room.muteAll.Load()
}

// ReplaceMutes replaces active room mute expirations.
func (room *Room) ReplaceMutes(mutes map[int64]time.Time) {
	room.mutex.Lock()
	room.muted = mutes
	room.mutex.Unlock()
}

// SetMuted stores or clears one active room mute expiration.
func (room *Room) SetMuted(playerID int64, endsAt time.Time) {
	room.mutex.Lock()
	if endsAt.IsZero() {
		delete(room.muted, playerID)
	} else {
		if room.muted == nil {
			room.muted = make(map[int64]time.Time)
		}
		room.muted[playerID] = endsAt
	}
	room.mutex.Unlock()
}
