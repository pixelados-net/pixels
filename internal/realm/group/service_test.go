package group

import (
	"context"
	"testing"

	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	groupruntime "github.com/niflaot/pixels/internal/realm/group/runtime"
)

// groupStore is a deterministic social membership fake.
type groupStore struct {
	// groupID stores the linked group.
	groupID int64
	// players stores group members.
	players []int64
	// found reports whether a group is linked.
	found bool
}

// snapshotStore provides player membership and inventory-link hydration fixtures.
type snapshotStore struct {
	grouprecord.Store
	// calls records persistence reads.
	calls *snapshotCalls
}

// snapshotCalls counts player snapshot persistence reads.
type snapshotCalls struct {
	// groups counts membership reads.
	groups int
	// links counts inventory-link reads.
	links int
}

// PlayerGroups returns one membership fixture.
func (store snapshotStore) PlayerGroups(context.Context, int64) ([]grouprecord.PlayerGroup, error) {
	store.calls.groups++
	return []grouprecord.PlayerGroup{{Group: grouprecord.Group{ID: 2}, Role: grouprecord.Member}}, nil
}

// PlayerInventoryFurnitureLinks returns one inventory link fixture.
func (store snapshotStore) PlayerInventoryFurnitureLinks(context.Context, int64) ([]grouprecord.GroupFurnitureLink, error) {
	store.calls.links++
	return []grouprecord.GroupFurnitureLink{{ItemID: 910128, GroupID: 2}}, nil
}

// TestPreparePlayerWarmsInventoryFurnitureLinks verifies reconnect restores group object data without a click query.
func TestPreparePlayerWarmsInventoryFurnitureLinks(t *testing.T) {
	cache := groupruntime.NewCache()
	calls := &snapshotCalls{}
	service := NewWiredService(snapshotStore{calls: calls}, cache)
	if err := service.PreparePlayer(context.Background(), 7); err != nil {
		t.Fatal(err)
	}
	if err := service.PreparePlayer(context.Background(), 7); err != nil {
		t.Fatal(err)
	}
	if groupID, found := cache.FurnitureGroup(910128); !found || groupID != 2 {
		t.Fatalf("groupID=%d found=%v", groupID, found)
	}
	if calls.groups != 1 || calls.links != 1 {
		t.Fatalf("groups=%d links=%d", calls.groups, calls.links)
	}
}

// RoomMembership returns the configured membership fixture.
func (store groupStore) RoomMembership(context.Context, int64) (int64, []int64, bool, error) {
	return store.groupID, store.players, store.found, nil
}

// TestServiceWarmsReadsAndReleasesSnapshots verifies no database work is needed by conditions.
func TestServiceWarmsReadsAndReleasesSnapshots(t *testing.T) {
	service := New(groupStore{groupID: 9, players: []int64{2, 4}, found: true})
	if member, loaded := service.IsRoomMember(7, 2); member || loaded {
		t.Fatal("cold room unexpectedly resolved")
	}
	if err := service.PrepareRoom(context.Background(), 7); err != nil {
		t.Fatal(err)
	}
	if member, loaded := service.IsRoomMember(7, 2); !member || !loaded {
		t.Fatalf("member=%v loaded=%v", member, loaded)
	}
	if member, loaded := service.IsRoomMember(7, 3); member || !loaded {
		t.Fatalf("non-member=%v loaded=%v", member, loaded)
	}
	service.CloseRoom(7)
	if _, loaded := service.IsRoomMember(7, 2); loaded {
		t.Fatal("closed room retained membership snapshot")
	}
}
