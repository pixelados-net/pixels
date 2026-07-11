package live

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	roomdoorbell "github.com/niflaot/pixels/internal/realm/room/doorbell"
)

// Room stores active runtime state for one loaded room.
type Room struct {
	// mutex protects active room state.
	mutex sync.RWMutex
	// snapshot stores stable room metadata.
	snapshot Snapshot
	// occupants stores occupants by player id.
	occupants map[int64]Occupant

	// rights stores persistent build-right holders while the room is active.
	rights map[int64]struct{}

	// muted stores active room mute expirations by player id.
	muted map[int64]time.Time

	// doorbell stores waiting entry requests and remains nil until first use.
	doorbell atomic.Pointer[roomdoorbell.Queue]

	// muteAll reports whether non-privileged room chat is globally muted.
	muteAll atomic.Bool

	// world stores loaded room world behavior.
	world *World

	// loadedAt stores when the active room was loaded.
	loadedAt time.Time

	// idleSince stores when the room became empty.
	idleSince *time.Time

	// closed reports whether the active room was closed.
	closed bool

	// loopCancel stops the room owner goroutine.
	loopCancel context.CancelFunc

	// loopDone closes when the room owner goroutine stops.
	loopDone chan struct{}
}

// NewRoom creates an active room.
func NewRoom(snapshot Snapshot) (*Room, error) {
	if !snapshot.Valid() {
		return nil, ErrInvalidRoom
	}

	return &Room{snapshot: snapshot, occupants: make(map[int64]Occupant), loadedAt: time.Now()}, nil
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
		room.world.addUnit(occupant.PlayerID)
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
		room.world.removeUnit(playerID)
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
	room.rights = nil
	room.muted = nil
	if room.world != nil {
		room.world.clearUnits()
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
	rights := make(map[int64]struct{}, len(playerIDs))
	for _, playerID := range playerIDs {
		if playerID > 0 {
			rights[playerID] = struct{}{}
		}
	}

	room.mutex.Lock()
	room.rights = rights
	room.mutex.Unlock()
}

// GrantRights adds a player to the active build-right projection.
func (room *Room) GrantRights(playerID int64) {
	if playerID <= 0 {
		return
	}

	room.mutex.Lock()
	if room.rights == nil {
		room.rights = make(map[int64]struct{})
	}
	room.rights[playerID] = struct{}{}
	room.mutex.Unlock()
}

// RevokeRights removes a player from the active build-right projection.
func (room *Room) RevokeRights(playerID int64) {
	room.mutex.Lock()
	delete(room.rights, playerID)
	room.mutex.Unlock()
}

// HasRights reports whether a player owns or holds active build rights.
func (room *Room) HasRights(playerID int64) bool {
	room.mutex.RLock()
	defer room.mutex.RUnlock()

	if playerID <= 0 {
		return false
	}
	if room.snapshot.OwnerPlayerID == playerID {
		return true
	}
	_, found := room.rights[playerID]

	return found
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
