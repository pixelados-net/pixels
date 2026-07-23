package runtime

import (
	"time"

	groupobservability "github.com/niflaot/pixels/internal/realm/group/observability"
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
)

func (cache *Cache) PutRoom(roomID int64, group GroupSnapshot, roles map[int64]grouprecord.Role) {
	roleCopy := make(map[int64]grouprecord.Role, len(roles))
	for playerID, role := range roles {
		roleCopy[playerID] = role
	}
	group.expiresAt = cache.expiry()
	cache.rooms.Store(roomID, &RoomSnapshot{RoomID: roomID, Group: group, roles: roleCopy})
	cache.generation.Add(1)
	cache.metrics.Record(groupobservability.SnapshotRefresh, groupobservability.KindRebind, groupobservability.Success)
}

// Room returns one warmed active-room generation.
func (cache *Cache) Room(roomID int64) (*RoomSnapshot, bool) {
	value, found := cache.rooms.Load(roomID)
	if !found {
		return nil, false
	}
	snapshot := value.(*RoomSnapshot)
	if expired(snapshot.Group.expiresAt) {
		if cache.rooms.CompareAndDelete(roomID, snapshot) {
			for _, itemID := range snapshot.furnitureIDs {
				cache.furniture.Delete(itemID)
			}
		}
		return nil, false
	}
	return snapshot, true
}

// CloseRoom releases one inactive room generation.
func (cache *Cache) CloseRoom(roomID int64) {
	if room, found := cache.Room(roomID); found {
		for _, itemID := range room.furnitureIDs {
			cache.furniture.Delete(itemID)
		}
	}
	cache.rooms.Delete(roomID)
	cache.generation.Add(1)
	cache.metrics.Record(groupobservability.SnapshotRefresh, groupobservability.KindDeactivate, groupobservability.Success)
}

// PutRoomFurniture replaces warmed furniture links for one active room.
func (cache *Cache) PutRoomFurniture(roomID int64, links []grouprecord.GroupFurnitureLink) {
	ids := make([]int64, len(links))
	for index, link := range links {
		ids[index] = link.ItemID
		cache.furniture.Store(link.ItemID, link.GroupID)
	}
	for {
		room, found := cache.Room(roomID)
		if !found {
			for _, itemID := range ids {
				cache.furniture.Delete(itemID)
			}
			return
		}
		replacement := *room
		replacement.furnitureIDs = ids
		if cache.rooms.CompareAndSwap(roomID, room, &replacement) {
			for _, itemID := range room.furnitureIDs {
				cache.furniture.Delete(itemID)
			}
			for index, link := range links {
				cache.furniture.Store(ids[index], link.GroupID)
			}
			break
		}
	}
	cache.generation.Add(1)
	cache.metrics.Record(groupobservability.SnapshotRefresh, groupobservability.KindDefault, groupobservability.Success)
}

// PutFurnitureLinkRecords warms a batch containing links from multiple groups.
func (cache *Cache) PutFurnitureLinkRecords(links []grouprecord.GroupFurnitureLink) {
	for _, link := range links {
		cache.furniture.Store(link.ItemID, link.GroupID)
	}
	if len(links) > 0 {
		cache.generation.Add(1)
	}
}

// FurnitureGroup returns one warmed item-to-group link without allocation.
func (cache *Cache) FurnitureGroup(itemID int64) (int64, bool) {
	value, found := cache.furniture.Load(itemID)
	if !found {
		return 0, false
	}
	return value.(int64), true
}

// IsRoomMember reports warmed room membership without allocation.
func (cache *Cache) IsRoomMember(roomID int64, playerID int64) (bool, bool) {
	room, loaded := cache.Room(roomID)
	if !loaded {
		return false, false
	}
	player, playerLoaded := cache.Player(playerID)
	if !playerLoaded {
		return false, false
	}
	_, member := player.Role(room.Group.Group.ID)
	return member, true
}

// RoleInRoom returns one warmed social role.
func (cache *Cache) RoleInRoom(roomID int64, playerID int64) (grouprecord.Role, bool, bool) {
	room, loaded := cache.Room(roomID)
	if !loaded {
		return 0, false, false
	}
	player, playerLoaded := cache.Player(playerID)
	if !playerLoaded {
		return 0, false, false
	}
	role, member := player.Role(room.Group.Group.ID)
	return role, member, true
}

// CanDecorate reports effective group decoration rights from warmed state.
func (cache *Cache) CanDecorate(roomID int64, playerID int64) (bool, bool) {
	role, member, loaded := cache.RoleInRoom(roomID, playerID)
	if !loaded {
		return false, false
	}
	if !member {
		return false, true
	}
	room, _ := cache.Room(roomID)
	return role == grouprecord.Owner || role == grouprecord.Admin || room.Group.Group.CanMembersDecorate, true
}

// Generation returns the current global invalidation generation.
func (cache *Cache) Generation() uint64 { return cache.generation.Load() }

// SetTTL configures safety-net snapshot expiration before serving requests.
func (cache *Cache) SetTTL(ttl time.Duration) { cache.ttl.Store(int64(ttl)) }

// expiry returns one configured absolute deadline or zero when disabled.
func (cache *Cache) expiry() int64 {
	ttl := cache.ttl.Load()
	if ttl <= 0 {
		return 0
	}
	return time.Now().Add(time.Duration(ttl)).UnixNano()
}

// expired reports whether one non-zero generation deadline elapsed.
func expired(deadline int64) bool { return deadline > 0 && time.Now().UnixNano() >= deadline }
