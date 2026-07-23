package runtime

import (
	"testing"
	"time"

	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
)

// TestSnapshotsResolveRolesFavoriteAndDecoration verifies immutable hot-path behavior.
func TestSnapshotsResolveRolesFavoriteAndDecoration(t *testing.T) {
	cache := NewCache()
	group := grouprecord.Group{ID: 9, Name: "Pixels", CanMembersDecorate: false}
	cache.PutGroup(GroupSnapshot{Group: group})
	cache.PutPlayer(7, []grouprecord.PlayerGroup{{Group: group, Role: grouprecord.Member, Favorite: true}})
	cache.PutPlayer(8, []grouprecord.PlayerGroup{{Group: group, Role: grouprecord.Admin}})
	cache.PutRoom(3, GroupSnapshot{Group: group}, nil)
	player, found := cache.Player(7)
	if !found {
		t.Fatal("missing player snapshot")
	}
	favorite, role, found := player.Favorite()
	if !found || favorite.ID != 9 || role != grouprecord.Member {
		t.Fatalf("unexpected favorite %#v %d", favorite, role)
	}
	if allowed, loaded := cache.CanDecorate(3, 7); allowed || !loaded {
		t.Fatalf("member decoration allowed=%v loaded=%v", allowed, loaded)
	}
	if allowed, loaded := cache.CanDecorate(3, 8); !allowed || !loaded {
		t.Fatalf("admin decoration allowed=%v loaded=%v", allowed, loaded)
	}
}

// TestSnapshotTTLRecoversFromMissedInvalidation verifies the safety-net expiry.
func TestSnapshotTTLRecoversFromMissedInvalidation(t *testing.T) {
	cache := NewCache()
	cache.SetTTL(time.Nanosecond)
	cache.PutGroup(GroupSnapshot{Group: grouprecord.Group{ID: 9}})
	cache.PutPlayer(7, []grouprecord.PlayerGroup{{Group: grouprecord.Group{ID: 9}, Role: grouprecord.Member}})
	time.Sleep(time.Millisecond)
	if _, found := cache.Group(9); found {
		t.Fatal("expired group generation remained visible")
	}
	if _, found := cache.Player(7); found {
		t.Fatal("expired player generation remained visible")
	}
}

// TestGroupMutationPropagatesAndDeactivationInvalidates verifies generation fan-out.
func TestGroupMutationPropagatesAndDeactivationInvalidates(t *testing.T) {
	cache := NewCache()
	group := grouprecord.Group{ID: 9, Name: "Before"}
	cache.PutGroup(GroupSnapshot{Group: group})
	cache.PutPlayer(7, []grouprecord.PlayerGroup{{Group: group, Role: grouprecord.Member, Favorite: true}})
	cache.PutRoom(3, GroupSnapshot{Group: group}, nil)
	group.Name = "After"
	cache.PutGroup(GroupSnapshot{Group: group})
	player, _ := cache.Player(7)
	room, _ := cache.Room(3)
	if player.Groups[0].Group.Name != "After" || room.Group.Group.Name != "After" {
		t.Fatalf("player=%#v room=%#v", player, room)
	}
	cache.DeleteGroup(9)
	player, _ = cache.Player(7)
	if len(player.Groups) != 0 || player.FavoriteID != 0 {
		t.Fatalf("player=%#v", player)
	}
	if _, found := cache.Room(3); found {
		t.Fatal("deactivated group retained room generation")
	}
}

// TestRoomFurnitureReplacementRemovesStaleLinks verifies room lifecycle cleanup.
func TestRoomFurnitureReplacementRemovesStaleLinks(t *testing.T) {
	cache := NewCache()
	cache.PutRoom(3, GroupSnapshot{Group: grouprecord.Group{ID: 9}}, nil)
	cache.PutRoomFurniture(3, []grouprecord.GroupFurnitureLink{{ItemID: 11, GroupID: 9}})
	if groupID, found := cache.FurnitureGroup(11); !found || groupID != 9 {
		t.Fatalf("group=%d found=%v", groupID, found)
	}
	cache.PutRoomFurniture(3, []grouprecord.GroupFurnitureLink{{ItemID: 12, GroupID: 9}})
	if _, found := cache.FurnitureGroup(11); found {
		t.Fatal("stale furniture link survived replacement")
	}
	cache.CloseRoom(3)
	if _, found := cache.FurnitureGroup(12); found {
		t.Fatal("room furniture link survived close")
	}
}

// BenchmarkIsRoomMember measures warmed WIRED membership lookup.
func BenchmarkIsRoomMember(b *testing.B) {
	cache := NewCache()
	group := grouprecord.Group{ID: 9}
	cache.PutPlayer(7, []grouprecord.PlayerGroup{{Group: group, Role: grouprecord.Member}})
	cache.PutRoom(3, GroupSnapshot{Group: group}, nil)
	b.ReportAllocs()
	for range b.N {
		_, _ = cache.IsRoomMember(3, 7)
	}
}

// BenchmarkRoleInRoom measures warmed room role lookup.
func BenchmarkRoleInRoom(b *testing.B) {
	cache := NewCache()
	group := grouprecord.Group{ID: 9}
	cache.PutPlayer(7, []grouprecord.PlayerGroup{{Group: group, Role: grouprecord.Admin}})
	cache.PutRoom(3, GroupSnapshot{Group: group}, nil)
	b.ReportAllocs()
	for range b.N {
		_, _, _ = cache.RoleInRoom(3, 7)
	}
}

// BenchmarkFavoriteProjection measures warmed favorite projection lookup.
func BenchmarkFavoriteProjection(b *testing.B) {
	cache := NewCache()
	cache.PutPlayer(7, []grouprecord.PlayerGroup{{Group: grouprecord.Group{ID: 9}, Role: grouprecord.Member, Favorite: true}})
	player, _ := cache.Player(7)
	b.ReportAllocs()
	for range b.N {
		_, _, _ = player.Favorite()
	}
}

// BenchmarkFurnitureGroup measures warmed group-furniture lookup.
func BenchmarkFurnitureGroup(b *testing.B) {
	cache := NewCache()
	cache.PutFurnitureLinks(9, []int64{11})
	b.ReportAllocs()
	for range b.N {
		_, _ = cache.FurnitureGroup(11)
	}
}

// TestPutFurnitureLinkRecordsWarmsMultipleGroups verifies one login batch preserves every group identity.
func TestPutFurnitureLinkRecordsWarmsMultipleGroups(t *testing.T) {
	cache := NewCache()
	cache.PutFurnitureLinkRecords([]grouprecord.GroupFurnitureLink{{ItemID: 11, GroupID: 9}, {ItemID: 12, GroupID: 10}})
	for itemID, expected := range map[int64]int64{11: 9, 12: 10} {
		if actual, found := cache.FurnitureGroup(itemID); !found || actual != expected {
			t.Fatalf("item=%d group=%d found=%v", itemID, actual, found)
		}
	}
}

// BenchmarkPlayerGroupsSnapshot measures warmed membership generation lookup.
func BenchmarkPlayerGroupsSnapshot(b *testing.B) {
	cache := NewCache()
	cache.PutPlayer(7, []grouprecord.PlayerGroup{{Group: grouprecord.Group{ID: 9}, Role: grouprecord.Member}})
	b.ReportAllocs()
	for range b.N {
		_, _ = cache.Player(7)
	}
}
