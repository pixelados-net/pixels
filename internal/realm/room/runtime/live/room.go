package live

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	roomdoorbell "github.com/niflaot/pixels/internal/realm/room/access/doorbell"
	rightsstate "github.com/niflaot/pixels/internal/realm/room/control/rights/state"
	roomowner "github.com/niflaot/pixels/internal/realm/room/runtime/live/owner"
	worldruntime "github.com/niflaot/pixels/internal/realm/room/world/runtime"
)

// Room stores active runtime state for one loaded room.
type Room struct {
	// mutex protects active room state.
	mutex sync.RWMutex
	// snapshot stores stable room metadata.
	snapshot Snapshot
	// occupants stores occupants by player id.
	occupants map[int64]Occupant
	// rights stores the active build-right projection.
	rights *rightsstate.State
	// muted stores active room mute expirations by player id.
	muted map[int64]time.Time
	// doorbell stores waiting entry requests and remains nil until first use.
	doorbell atomic.Pointer[roomdoorbell.Queue]
	// muteAll reports whether non-privileged room chat is globally muted.
	muteAll atomic.Bool
	// world stores loaded room world behavior.
	world *worldruntime.World
	// loadedAt stores when the active room was loaded.
	loadedAt time.Time
	// idleSince stores when the room became empty.
	idleSince *time.Time
	// closed reports whether the active room was closed.
	closed bool
	// loop owns the room's periodic movement and doorbell cycle.
	loop roomowner.Loop
}

// NewRoom creates an active room.
func NewRoom(snapshot Snapshot) (*Room, error) {
	if !snapshot.Valid() {
		return nil, ErrInvalidRoom
	}

	return &Room{
		snapshot: snapshot, occupants: make(map[int64]Occupant),
		rights: rightsstate.New(snapshot.OwnerPlayerID), loadedAt: time.Now(),
	}, nil
}

// ID returns the room id.
func (room *Room) ID() int64 {
	room.mutex.RLock()
	defer room.mutex.RUnlock()

	return room.snapshot.ID
}

// Snapshot returns active room metadata.
func (room *Room) Snapshot() Snapshot {
	room.mutex.RLock()
	defer room.mutex.RUnlock()

	return room.snapshot
}

// Join adds or replaces an active room occupant.
func (room *Room) Join(occupant Occupant) (Occupancy, error) {
	return room.JoinWithCapacity(occupant, false)
}

// JoinWithCapacity adds an occupant with an optional capacity bypass.
func (room *Room) JoinWithCapacity(occupant Occupant, bypassCapacity bool) (Occupancy, error) {
	if !occupant.Valid() {
		return Occupancy{}, ErrInvalidOccupant
	}
	room.mutex.Lock()
	defer room.mutex.Unlock()
	if room.closed {
		return Occupancy{}, ErrRoomClosed
	}
	if _, exists := room.occupants[occupant.PlayerID]; !exists && len(room.occupants) >= room.snapshot.MaxUsers && !bypassCapacity {
		return Occupancy{}, ErrRoomFull
	}
	room.occupants[occupant.PlayerID] = occupant.WithJoinTime(time.Now())
	if room.world != nil {
		room.world.AddUnit(occupant.PlayerID)
	}
	room.idleSince = nil

	return room.occupancyLocked(), nil
}

// Leave removes an active room occupant.
func (room *Room) Leave(playerID int64) (Occupancy, bool) {
	room.mutex.Lock()
	defer room.mutex.Unlock()
	if _, found := room.occupants[playerID]; !found {
		return Occupancy{}, false
	}
	delete(room.occupants, playerID)
	if room.world != nil {
		room.world.RemoveUnit(playerID)
	}
	if len(room.occupants) == 0 {
		now := time.Now()
		room.idleSince = &now
	}

	return room.occupancyLocked(), true
}

// Occupancy returns a stable occupancy snapshot.
func (room *Room) Occupancy() Occupancy {
	room.mutex.RLock()
	defer room.mutex.RUnlock()

	return room.occupancyLocked()
}

// Occupants returns a stable occupant snapshot.
func (room *Room) Occupants() []Occupant {
	room.mutex.RLock()
	defer room.mutex.RUnlock()
	occupants := make([]Occupant, 0, len(room.occupants))
	for _, occupant := range room.occupants {
		occupants = append(occupants, occupant)
	}

	return occupants
}

// Close marks the active room as closed.
func (room *Room) Close() Occupancy {
	occupancy, _ := room.CloseWithDoorbell()

	return occupancy
}

// CloseWithDoorbell closes the room and returns every pending entry request.
func (room *Room) CloseWithDoorbell() (Occupancy, []roomdoorbell.Expired) {
	room.stopLoop()
	room.mutex.Lock()
	room.closed = true
	room.occupants = make(map[int64]Occupant)
	room.rights.Clear()
	room.muted = nil
	if room.world != nil {
		room.world.ClearUnits()
	}
	now := time.Now()
	room.idleSince = &now
	queue := room.doorbell.Swap(nil)
	occupancy := room.occupancyLocked()
	room.mutex.Unlock()
	if queue == nil {
		return occupancy, nil
	}

	return occupancy, queue.Drain(roomdoorbell.ExpiredRoomClosed)
}

// ReplaceRights replaces the active room build-right projection.
func (room *Room) ReplaceRights(playerIDs []int64) {
	room.rights.ReplaceRights(playerIDs)
}

// GrantRights adds a player to the active build-right projection.
func (room *Room) GrantRights(playerID int64) {
	room.rights.GrantRights(playerID)
}

// RevokeRights removes a player from the active build-right projection.
func (room *Room) RevokeRights(playerID int64) {
	room.rights.RevokeRights(playerID)
}

// HasRights reports whether a player owns or holds active build rights.
func (room *Room) HasRights(playerID int64) bool {
	return room.rights.HasRights(playerID)
}

// CanManageFurniture reports whether a player may manage room furniture.
func (room *Room) CanManageFurniture(playerID int64) bool {
	return room.HasRights(playerID)
}

// IdleSince returns when the room became empty.
func (room *Room) IdleSince() *time.Time {
	room.mutex.RLock()
	defer room.mutex.RUnlock()
	if room.idleSince == nil {
		return nil
	}
	idleSince := *room.idleSince

	return &idleSince
}

// occupancyLocked returns occupancy while a room lock is held.
func (room *Room) occupancyLocked() Occupancy {
	playerIDs := make([]int64, 0, len(room.occupants))
	for playerID := range room.occupants {
		playerIDs = append(playerIDs, playerID)
	}

	return Occupancy{RoomID: room.snapshot.ID, CategoryID: room.snapshot.CategoryID, Count: len(room.occupants), MaxUsers: room.snapshot.MaxUsers, PlayerIDs: playerIDs}
}

// startLoop starts the room owner goroutine.
func (room *Room) startLoop(ctx context.Context, interval time.Duration, movementPublisher MovementPublisher, doorbellPublisher DoorbellPublisher, cyclePublisher CyclePublisher, doorbellTimeout time.Duration) {
	if movementPublisher == nil && doorbellPublisher == nil && cyclePublisher == nil {
		return
	}
	room.loop.Start(ctx, interval, func(loopCtx context.Context) {
		movements := room.Tick()
		if len(movements) > 0 && movementPublisher != nil {
			_ = movementPublisher(loopCtx, room, movements)
		}
		expired := room.SweepDoorbell(time.Now(), doorbellTimeout)
		if len(expired) > 0 && doorbellPublisher != nil {
			_ = doorbellPublisher(loopCtx, room, expired)
		}
		if cyclePublisher != nil {
			_ = cyclePublisher(loopCtx, room, time.Now())
		}
	})
}

// stopLoop stops the room owner goroutine.
func (room *Room) stopLoop() {
	room.loop.Stop()
}
