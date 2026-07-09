package live

import (
	"context"
	"sort"
	"sync"
	"time"

	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
)

// TileHeight describes one resolved tile's current walkable height and stacking state.
type TileHeight struct {
	// Valid reports whether the tile is part of the room.
	Valid bool

	// Height stores the current walkable top height.
	Height grid.Height

	// StackingBlocked reports whether new items cannot stack on this tile.
	StackingBlocked bool
}

// Room stores active runtime state for one loaded room.
type Room struct {
	// mutex protects active room state.
	mutex sync.RWMutex

	// snapshot stores stable room metadata.
	snapshot Snapshot

	// occupants stores occupants by player id.
	occupants map[int64]Occupant

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
	if !occupant.Valid() {
		return Occupancy{}, ErrInvalidOccupant
	}

	room.mutex.Lock()
	defer room.mutex.Unlock()

	if room.closed {
		return Occupancy{}, ErrRoomClosed
	}
	if _, exists := room.occupants[occupant.PlayerID]; !exists && len(room.occupants) >= room.snapshot.MaxUsers {
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
	room.stopLoop()

	room.mutex.Lock()
	defer room.mutex.Unlock()

	room.closed = true
	room.occupants = make(map[int64]Occupant)
	if room.world != nil {
		room.world.clearUnits()
	}
	now := time.Now()
	room.idleSince = &now

	return room.occupancyLocked()
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

// Units returns stable world unit snapshots.
func (room *Room) Units() []UnitSnapshot {
	room.mutex.RLock()
	defer room.mutex.RUnlock()

	if room.world == nil {
		return nil
	}

	playerIDs := make([]int64, 0, len(room.world.units))
	for playerID := range room.world.units {
		playerIDs = append(playerIDs, playerID)
	}
	sort.Slice(playerIDs, func(left int, right int) bool {
		return playerIDs[left] < playerIDs[right]
	})

	units := make([]UnitSnapshot, 0, len(playerIDs))
	for _, playerID := range playerIDs {
		roomUnit := room.world.units[playerID]
		units = append(units, unitSnapshot(playerID, roomUnit))
	}

	return units
}

// FurnitureItems returns stable placed furniture item snapshots.
func (room *Room) FurnitureItems() []worldfurniture.Item {
	room.mutex.RLock()
	defer room.mutex.RUnlock()

	if room.world == nil {
		return nil
	}

	ids := make([]int64, 0, len(room.world.furniture))
	for id := range room.world.furniture {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(left int, right int) bool {
		return ids[left] < ids[right]
	})

	items := make([]worldfurniture.Item, 0, len(ids))
	for _, id := range ids {
		items = append(items, room.world.furniture[id])
	}

	return items
}

// SurfaceHeights returns the room's current per-tile walkable heights, row-major by (y*width+x).
func (room *Room) SurfaceHeights() (uint16, uint16, []TileHeight) {
	room.mutex.RLock()
	defer room.mutex.RUnlock()

	if room.world == nil {
		return 0, 0, nil
	}

	width, height := room.world.grid.Width(), room.world.grid.Height()
	tiles := make([]TileHeight, 0, int(width)*int(height))
	for y := uint16(0); y < height; y++ {
		for x := uint16(0); x < width; x++ {
			section, err := room.world.resolver.TopSection(grid.Point{X: x, Y: y})
			if err != nil {
				tiles = append(tiles, TileHeight{})

				continue
			}
			tiles = append(tiles, TileHeight{Valid: true, Height: section.Z(), StackingBlocked: !section.Stacking()})
		}
	}

	return width, height, tiles
}

// SlotOccupant returns a player id currently occupying any sit/lay slot of a placed furniture item.
func (room *Room) SlotOccupant(itemID int64) (int64, bool) {
	room.mutex.RLock()
	defer room.mutex.RUnlock()

	if room.world == nil {
		return 0, false
	}
	item, found := room.world.furniture[itemID]
	if !found {
		return 0, false
	}

	for _, slot := range worldfurniture.Slots(item) {
		if playerID, occupied := room.world.slotOccupants[slot.Point]; occupied {
			return playerID, true
		}
	}

	return 0, false
}

// unitSnapshot maps a world unit to a runtime snapshot.
func unitSnapshot(playerID int64, roomUnit *worldunit.Unit) UnitSnapshot {
	return UnitSnapshot{
		PlayerID:     playerID,
		UnitID:       roomUnit.ID(),
		Position:     roomUnit.Position(),
		Previous:     roomUnit.Previous(),
		BodyRotation: roomUnit.BodyRotation(),
		HeadRotation: roomUnit.HeadRotation(),
		Moving:       roomUnit.Moving(),
		Statuses:     roomUnit.Statuses(),
	}
}
