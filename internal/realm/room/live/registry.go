package live

import (
	"context"
	"sync"
	"time"
)

// OccupancyPublisher publishes occupancy changes.
type OccupancyPublisher func(context.Context, Occupancy) error

// Registry stores active rooms.
type Registry struct {
	// mutex protects active room indexes.
	mutex sync.RWMutex

	// rooms stores active rooms by room id.
	rooms map[int64]*Room

	// byPlayer stores active room ids by player id.
	byPlayer map[int64]int64

	// publish publishes occupancy changes.
	publish OccupancyPublisher
}

// NewRegistry creates an active room registry.
func NewRegistry(publisher OccupancyPublisher) *Registry {
	return &Registry{rooms: make(map[int64]*Room), byPlayer: make(map[int64]int64), publish: publisher}
}

// Activate registers an active room.
func (registry *Registry) Activate(snapshot Snapshot) (*Room, error) {
	room, err := NewRoom(snapshot)
	if err != nil {
		return nil, err
	}

	registry.mutex.Lock()
	defer registry.mutex.Unlock()

	if active, found := registry.rooms[snapshot.ID]; found {
		return active, nil
	}

	registry.rooms[snapshot.ID] = room

	return room, nil
}

// Find returns an active room by id.
func (registry *Registry) Find(roomID int64) (*Room, bool) {
	registry.mutex.RLock()
	defer registry.mutex.RUnlock()

	room, found := registry.rooms[roomID]

	return room, found
}

// Join adds a player to an active room.
func (registry *Registry) Join(ctx context.Context, roomID int64, occupant Occupant) (Occupancy, error) {
	if !occupant.Valid() {
		return Occupancy{}, ErrInvalidOccupant
	}

	room, found := registry.Find(roomID)
	if !found {
		return Occupancy{}, ErrRoomNotFound
	}

	registry.removeIndexedPlayer(ctx, occupant.PlayerID)
	occupancy, err := room.Join(occupant)
	if err != nil {
		return Occupancy{}, err
	}

	registry.mutex.Lock()
	registry.byPlayer[occupant.PlayerID] = roomID
	registry.mutex.Unlock()

	return occupancy, registry.publishOccupancy(ctx, occupancy)
}

// Leave removes a player from its active room.
func (registry *Registry) Leave(ctx context.Context, playerID int64) (Occupancy, bool, error) {
	registry.mutex.RLock()
	roomID, found := registry.byPlayer[playerID]
	room := registry.rooms[roomID]
	registry.mutex.RUnlock()
	if !found || room == nil {
		return Occupancy{}, false, nil
	}

	occupancy, removed := room.Leave(playerID)
	if !removed {
		return Occupancy{}, false, nil
	}

	registry.mutex.Lock()
	delete(registry.byPlayer, playerID)
	registry.mutex.Unlock()

	return occupancy, true, registry.publishOccupancy(ctx, occupancy)
}

// RemovePlayer removes a player from its active room.
func (registry *Registry) RemovePlayer(ctx context.Context, playerID int64) (Occupancy, bool, error) {
	return registry.Leave(ctx, playerID)
}

// Close closes and unregisters an active room.
func (registry *Registry) Close(ctx context.Context, roomID int64) (Occupancy, bool, error) {
	registry.mutex.Lock()
	room, found := registry.rooms[roomID]
	if !found {
		registry.mutex.Unlock()
		return Occupancy{}, false, nil
	}
	delete(registry.rooms, roomID)
	for playerID, activeRoomID := range registry.byPlayer {
		if activeRoomID == roomID {
			delete(registry.byPlayer, playerID)
		}
	}
	registry.mutex.Unlock()

	occupancy := room.Close()

	return occupancy, true, registry.publishOccupancy(ctx, occupancy)
}

// UnloadIdle closes rooms that have been empty for at least idleFor.
func (registry *Registry) UnloadIdle(ctx context.Context, idleFor time.Duration, now time.Time) ([]Occupancy, error) {
	registry.mutex.RLock()
	roomIDs := make([]int64, 0, len(registry.rooms))
	for roomID, room := range registry.rooms {
		idleSince := room.IdleSince()
		if idleSince != nil && !idleSince.Add(idleFor).After(now) {
			roomIDs = append(roomIDs, roomID)
		}
	}
	registry.mutex.RUnlock()

	closedOccupancies := make([]Occupancy, 0, len(roomIDs))
	for _, roomID := range roomIDs {
		occupancy, closed, err := registry.Close(ctx, roomID)
		if err != nil {
			return closedOccupancies, err
		}
		if closed {
			closedOccupancies = append(closedOccupancies, occupancy)
		}
	}

	return closedOccupancies, nil
}

// Count returns the number of active rooms.
func (registry *Registry) Count() int {
	registry.mutex.RLock()
	defer registry.mutex.RUnlock()

	return len(registry.rooms)
}

// Snapshot returns stable active room references.
func (registry *Registry) Snapshot() []*Room {
	registry.mutex.RLock()
	defer registry.mutex.RUnlock()

	rooms := make([]*Room, 0, len(registry.rooms))
	for _, room := range registry.rooms {
		rooms = append(rooms, room)
	}

	return rooms
}

// removeIndexedPlayer removes a player from any indexed room.
func (registry *Registry) removeIndexedPlayer(ctx context.Context, playerID int64) {
	if playerID <= 0 {
		return
	}

	_, _, _ = registry.Leave(ctx, playerID)
}

// publishOccupancy publishes occupancy when configured.
func (registry *Registry) publishOccupancy(ctx context.Context, occupancy Occupancy) error {
	if registry.publish == nil {
		return nil
	}

	return registry.publish(ctx, occupancy)
}
