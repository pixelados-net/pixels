package group

import (
	"context"
	"sync"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	groupruntime "github.com/niflaot/pixels/internal/realm/group/runtime"
)

// roomSnapshot stores one immutable room group membership set.
type roomSnapshot struct {
	// groupID identifies the linked social group.
	groupID int64
	// members stores current member IDs.
	members map[int64]struct{}
}

// Service caches room group membership outside WIRED hot paths.
type Service struct {
	// store reads durable membership.
	store Store
	// records reads bounded social-group snapshots in production.
	records grouprecord.Store
	// cache stores shared immutable group/player/room generations.
	cache *groupruntime.Cache
	// rooms stores immutable prepared snapshots.
	rooms sync.Map
}

// New creates a social group read service.
func New(store Store) *Service { return &Service{store: store} }

// NewWiredService creates the production WIRED facade over shared immutable snapshots.
func NewWiredService(records grouprecord.Store, cache *groupruntime.Cache) *Service {
	return &Service{records: records, cache: cache}
}

// PrepareRoom refreshes one room's group membership snapshot.
func (service *Service) PrepareRoom(ctx context.Context, roomID int64) error {
	if service.records != nil && service.cache != nil {
		group, found, err := service.records.RoomGroup(ctx, roomID)
		if err != nil {
			return err
		}
		if !found {
			service.cache.CloseRoom(roomID)
			return nil
		}
		parts, err := service.records.BadgeParts(ctx, group.ID)
		if err != nil {
			return err
		}
		snapshot := groupruntime.GroupSnapshot{Group: group, BadgeParts: parts}
		service.cache.PutGroup(snapshot)
		service.cache.PutRoom(roomID, snapshot, nil)
		links, err := service.records.RoomFurnitureLinks(ctx, roomID)
		if err != nil {
			return err
		}
		service.cache.PutRoomFurniture(roomID, links)
		return nil
	}
	groupID, players, found, err := service.store.RoomMembership(ctx, roomID)
	if err != nil {
		return err
	}
	if !found {
		service.rooms.Delete(roomID)
		return nil
	}
	members := make(map[int64]struct{}, len(players))
	for _, playerID := range players {
		members[playerID] = struct{}{}
	}
	service.rooms.Store(roomID, &roomSnapshot{groupID: groupID, members: members})
	return nil
}

// PreparePlayer warms one authenticated player's sorted membership generation.
func (service *Service) PreparePlayer(ctx context.Context, playerID int64) error {
	if service.records == nil || service.cache == nil {
		return nil
	}
	if _, found := service.cache.Player(playerID); found {
		return nil
	}
	groups, err := service.records.PlayerGroups(ctx, playerID)
	if err != nil {
		return err
	}
	links, err := service.records.PlayerInventoryFurnitureLinks(ctx, playerID)
	if err != nil {
		return err
	}
	service.cache.PutPlayer(playerID, groups)
	service.cache.PutFurnitureLinkRecords(links)
	return nil
}

// IsRoomMember reports membership from a warmed immutable snapshot.
func (service *Service) IsRoomMember(roomID int64, playerID int64) (bool, bool) {
	if service.cache != nil {
		return service.cache.IsRoomMember(roomID, playerID)
	}
	value, found := service.rooms.Load(roomID)
	if !found {
		return false, false
	}
	_, member := value.(*roomSnapshot).members[playerID]
	return member, true
}

// CanDecorate reports warmed social-group decoration rights.
func (service *Service) CanDecorate(roomID int64, playerID int64) bool {
	if service.cache == nil {
		return false
	}
	allowed, loaded := service.cache.CanDecorate(roomID, playerID)
	return loaded && allowed
}

// Favorite returns warmed favorite unit projection data without allocation.
func (service *Service) Favorite(playerID int64) (int64, int32, string) {
	if service.cache == nil {
		return -1, -1, ""
	}
	player, found := service.cache.Player(playerID)
	if !found {
		return -1, -1, ""
	}
	group, role, found := player.Favorite()
	if !found {
		return -1, -1, ""
	}
	return group.ID, int32(role), group.Name
}

// Furniture returns warmed linked group furniture identity without database access.
func (service *Service) Furniture(itemID int64) (furnituremodel.GroupData, bool) {
	if service.cache == nil {
		return furnituremodel.GroupData{}, false
	}
	groupID, found := service.cache.FurnitureGroup(itemID)
	if !found {
		return furnituremodel.GroupData{}, false
	}
	group, found := service.cache.Group(groupID)
	if !found {
		return furnituremodel.GroupData{}, false
	}
	return furnituremodel.GroupData{GroupID: groupID, BadgeCode: group.Group.BadgeCode, ColorAHex: group.Group.ColorAHex, ColorBHex: group.Group.ColorBHex}, true
}

// RoomGroups returns one bounded group map with one batch query on cache misses.
func (service *Service) RoomGroups(ctx context.Context, roomIDs []int64) (map[int64]grouprecord.Group, error) {
	result := make(map[int64]grouprecord.Group, len(roomIDs))
	missing := make([]int64, 0, len(roomIDs))
	for _, roomID := range roomIDs {
		if room, found := service.cache.Room(roomID); found {
			result[roomID] = room.Group.Group
		} else {
			missing = append(missing, roomID)
		}
	}
	bindings, err := service.records.RoomGroups(ctx, missing)
	if err != nil {
		return nil, err
	}
	for _, binding := range bindings {
		result[binding.RoomID] = binding.Group
		service.cache.PutGroup(groupruntime.GroupSnapshot{Group: binding.Group})
	}
	return result, nil
}

// PlayerGroups returns one player's active group memberships from the warmed generation.
func (service *Service) PlayerGroups(ctx context.Context, playerID int64) ([]grouprecord.PlayerGroup, error) {
	if err := service.PreparePlayer(ctx, playerID); err != nil {
		return nil, err
	}
	snapshot, found := service.cache.Player(playerID)
	if !found {
		return []grouprecord.PlayerGroup{}, nil
	}
	return append([]grouprecord.PlayerGroup(nil), snapshot.Groups...), nil
}

// PopularGroups returns active groups ordered by descending member count.
func (service *Service) PopularGroups(ctx context.Context, limit int) ([]grouprecord.Group, error) {
	if service.records == nil {
		return []grouprecord.Group{}, nil
	}
	return service.records.PopularGroups(ctx, limit)
}

// Group returns one active social group by identifier.
func (service *Service) Group(ctx context.Context, groupID int64) (grouprecord.Group, bool, error) {
	if snapshot, found := service.cache.Group(groupID); found {
		return snapshot.Group, true, nil
	}
	group, found, err := service.records.Group(ctx, groupID, false)
	if err == nil && found {
		service.cache.PutGroup(groupruntime.GroupSnapshot{Group: group})
	}
	return group, found, err
}

// RoomGroupInfo returns one viewer-specific room group without a roster query.
func (service *Service) RoomGroupInfo(ctx context.Context, roomID int64, playerID int64) (grouprecord.Group, bool, bool, error) {
	if room, found := service.cache.Room(roomID); found {
		player, loaded := service.cache.Player(playerID)
		if !loaded {
			return room.Group.Group, false, true, nil
		}
		_, member := player.Role(room.Group.Group.ID)
		return room.Group.Group, member, true, nil
	}
	group, found, err := service.records.RoomGroup(ctx, roomID)
	if err != nil || !found {
		return grouprecord.Group{}, false, found, err
	}
	service.cache.PutGroup(groupruntime.GroupSnapshot{Group: group})
	if err = service.PreparePlayer(ctx, playerID); err != nil {
		return grouprecord.Group{}, false, false, err
	}
	player, _ := service.cache.Player(playerID)
	_, member := player.Role(group.ID)
	return group, member, true, nil
}

// CloseRoom releases one inactive room snapshot.
func (service *Service) CloseRoom(roomID int64) {
	if service.cache != nil {
		service.cache.CloseRoom(roomID)
	}
	service.rooms.Delete(roomID)
}
