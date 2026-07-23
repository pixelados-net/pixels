// Package runtime owns immutable social-group snapshots used by hot paths.
package runtime

import (
	"sort"
	"sync"
	"sync/atomic"

	groupobservability "github.com/niflaot/pixels/internal/realm/group/observability"
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
)

// GroupSnapshot stores immutable group metadata and badge parts.
type GroupSnapshot struct {
	// Group stores durable metadata.
	Group grouprecord.Group
	// BadgeParts stores normalized layers.
	BadgeParts []grouprecord.BadgePart
	// expiresAt stores the safety-net invalidation deadline in Unix nanoseconds.
	expiresAt int64
}

// PlayerSnapshot stores one immutable sorted membership generation.
type PlayerSnapshot struct {
	// Groups stores active memberships sorted by group identifier.
	Groups []grouprecord.PlayerGroup
	// FavoriteID identifies the favorite group or zero.
	FavoriteID int64
	// expiresAt stores the safety-net invalidation deadline in Unix nanoseconds.
	expiresAt int64
}

// Role returns one membership role without allocating.
func (snapshot *PlayerSnapshot) Role(groupID int64) (grouprecord.Role, bool) {
	if snapshot == nil {
		return 0, false
	}
	index := sort.Search(len(snapshot.Groups), func(index int) bool { return snapshot.Groups[index].Group.ID >= groupID })
	if index >= len(snapshot.Groups) || snapshot.Groups[index].Group.ID != groupID {
		return 0, false
	}
	return snapshot.Groups[index].Role, true
}

// Favorite returns favorite projection data without allocating.
func (snapshot *PlayerSnapshot) Favorite() (grouprecord.Group, grouprecord.Role, bool) {
	if snapshot == nil || snapshot.FavoriteID <= 0 {
		return grouprecord.Group{}, 0, false
	}
	index := sort.Search(len(snapshot.Groups), func(index int) bool { return snapshot.Groups[index].Group.ID >= snapshot.FavoriteID })
	if index >= len(snapshot.Groups) || snapshot.Groups[index].Group.ID != snapshot.FavoriteID {
		return grouprecord.Group{}, 0, false
	}
	return snapshot.Groups[index].Group, snapshot.Groups[index].Role, true
}

// RoomSnapshot stores immutable room group and relevant role lookup state.
type RoomSnapshot struct {
	// RoomID identifies the room.
	RoomID int64
	// Group stores linked group metadata.
	Group GroupSnapshot
	// roles stores active member roles.
	roles map[int64]grouprecord.Role
	// furnitureIDs stores linked item identifiers released with the room.
	furnitureIDs []int64
}

// Role returns a warmed room membership role without allocating.
func (snapshot *RoomSnapshot) Role(playerID int64) (grouprecord.Role, bool) {
	if snapshot == nil {
		return 0, false
	}
	role, found := snapshot.roles[playerID]
	return role, found
}

// Cache stores immutable group, player, and active-room generations.
type Cache struct {
	// groups stores group snapshots by identifier.
	groups sync.Map
	// players stores player snapshots by identifier.
	players sync.Map
	// rooms stores active room snapshots by identifier.
	rooms sync.Map
	// furniture stores warmed item-to-group links.
	furniture sync.Map
	// generation stores the global invalidation generation.
	generation atomic.Uint64
	// ttl stores the safety-net generation lifetime in nanoseconds.
	ttl atomic.Int64
	// metrics stores bounded process-wide group telemetry.
	metrics *groupobservability.Metrics
}

// NewCache creates an empty snapshot cache.
func NewCache() *Cache { return &Cache{} }

// SetMetrics attaches process-wide telemetry before serving requests.
func (cache *Cache) SetMetrics(metrics *groupobservability.Metrics) { cache.metrics = metrics }

// PutGroup replaces one immutable group generation.
func (cache *Cache) PutGroup(snapshot GroupSnapshot) {
	snapshot.expiresAt = cache.expiry()
	parts := append([]grouprecord.BadgePart(nil), snapshot.BadgeParts...)
	snapshot.BadgeParts = parts
	cache.groups.Store(snapshot.Group.ID, &snapshot)
	cache.rooms.Range(func(key, value any) bool {
		for room := value.(*RoomSnapshot); room.Group.Group.ID == snapshot.Group.ID; {
			replacement := *room
			replacement.Group = snapshot
			if cache.rooms.CompareAndSwap(key, room, &replacement) {
				break
			}
			current, found := cache.rooms.Load(key)
			if !found {
				break
			}
			room = current.(*RoomSnapshot)
		}
		return true
	})
	cache.players.Range(func(key, value any) bool {
		for player := value.(*PlayerSnapshot); ; {
			index := sort.Search(len(player.Groups), func(index int) bool { return player.Groups[index].Group.ID >= snapshot.Group.ID })
			if index >= len(player.Groups) || player.Groups[index].Group.ID != snapshot.Group.ID {
				break
			}
			groups := append([]grouprecord.PlayerGroup(nil), player.Groups...)
			groups[index].Group = snapshot.Group
			if cache.players.CompareAndSwap(key, player, &PlayerSnapshot{Groups: groups, FavoriteID: player.FavoriteID, expiresAt: player.expiresAt}) {
				break
			}
			current, found := cache.players.Load(key)
			if !found {
				break
			}
			player = current.(*PlayerSnapshot)
		}
		return true
	})
	cache.generation.Add(1)
	cache.metrics.Record(groupobservability.SnapshotRefresh, groupobservability.KindUpdate, groupobservability.Success)
}

// Group returns one warmed group generation.
func (cache *Cache) Group(groupID int64) (*GroupSnapshot, bool) {
	value, found := cache.groups.Load(groupID)
	if !found {
		return nil, false
	}
	snapshot := value.(*GroupSnapshot)
	if expired(snapshot.expiresAt) {
		cache.groups.CompareAndDelete(groupID, snapshot)
		return nil, false
	}
	return snapshot, true
}

// DeleteGroup invalidates one group generation.
func (cache *Cache) DeleteGroup(groupID int64) {
	cache.groups.Delete(groupID)
	cache.furniture.Range(func(key, value any) bool {
		if value.(int64) == groupID {
			cache.furniture.Delete(key)
		}
		return true
	})
	cache.rooms.Range(func(key, value any) bool {
		room := value.(*RoomSnapshot)
		if room.Group.Group.ID == groupID {
			cache.CloseRoom(key.(int64))
		}
		return true
	})
	cache.players.Range(func(key, value any) bool {
		for player := value.(*PlayerSnapshot); ; {
			index := sort.Search(len(player.Groups), func(index int) bool { return player.Groups[index].Group.ID >= groupID })
			if index >= len(player.Groups) || player.Groups[index].Group.ID != groupID {
				break
			}
			groups := make([]grouprecord.PlayerGroup, 0, len(player.Groups)-1)
			groups = append(groups, player.Groups[:index]...)
			groups = append(groups, player.Groups[index+1:]...)
			favoriteID := player.FavoriteID
			if favoriteID == groupID {
				favoriteID = 0
			}
			if cache.players.CompareAndSwap(key, player, &PlayerSnapshot{Groups: groups, FavoriteID: favoriteID, expiresAt: player.expiresAt}) {
				break
			}
			current, found := cache.players.Load(key)
			if !found {
				break
			}
			player = current.(*PlayerSnapshot)
		}
		return true
	})
	cache.generation.Add(1)
	cache.metrics.Record(groupobservability.SnapshotRefresh, groupobservability.KindDeactivate, groupobservability.Success)
}

// PutFurnitureLinks warms inventory or placed item links after commit.
func (cache *Cache) PutFurnitureLinks(groupID int64, itemIDs []int64) {
	for _, itemID := range itemIDs {
		cache.furniture.Store(itemID, groupID)
	}
	if len(itemIDs) > 0 {
		cache.generation.Add(1)
	}
}

// PutPlayer replaces one immutable sorted player generation.
func (cache *Cache) PutPlayer(playerID int64, groups []grouprecord.PlayerGroup) {
	items := append([]grouprecord.PlayerGroup(nil), groups...)
	sort.Slice(items, func(left, right int) bool { return items[left].Group.ID < items[right].Group.ID })
	favoriteID := int64(0)
	for _, item := range items {
		if item.Favorite {
			favoriteID = item.Group.ID
			break
		}
	}
	cache.players.Store(playerID, &PlayerSnapshot{Groups: items, FavoriteID: favoriteID, expiresAt: cache.expiry()})
	cache.generation.Add(1)
	cache.metrics.Record(groupobservability.SnapshotRefresh, groupobservability.KindFavorite, groupobservability.Success)
}

// Player returns one warmed player generation.
func (cache *Cache) Player(playerID int64) (*PlayerSnapshot, bool) {
	value, found := cache.players.Load(playerID)
	if !found {
		return nil, false
	}
	snapshot := value.(*PlayerSnapshot)
	if expired(snapshot.expiresAt) {
		cache.players.CompareAndDelete(playerID, snapshot)
		return nil, false
	}
	return snapshot, true
}

// DeletePlayer invalidates one player generation.
func (cache *Cache) DeletePlayer(playerID int64) {
	cache.players.Delete(playerID)
	cache.generation.Add(1)
}

// PutRoom replaces one immutable active-room generation.
