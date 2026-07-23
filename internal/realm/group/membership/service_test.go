package membership

import (
	"context"
	"errors"
	"testing"

	groupconfig "github.com/niflaot/pixels/internal/realm/group/config"
	changedevent "github.com/niflaot/pixels/internal/realm/group/membership/events/changed"
	grouprecord "github.com/niflaot/pixels/internal/realm/group/record"
	groupruntime "github.com/niflaot/pixels/internal/realm/group/runtime"
	"github.com/niflaot/pixels/pkg/bus"
)

// membershipStore supplies deterministic membership persistence behavior.
type membershipStore struct {
	// Store supplies unused persistence methods.
	grouprecord.Store
	// group stores deterministic group metadata.
	group grouprecord.Group
	// members stores deterministic active roles.
	members map[int64]grouprecord.Membership
	// pending stores deterministic membership requests.
	pending map[int64]bool
	// playerGroups stores one player projection.
	playerGroups []grouprecord.PlayerGroup
	// memberPageLevel stores the latest roster filter passed to persistence.
	memberPageLevel int32
	// favorite stores the latest preference mutation.
	favorite *int64
	// linked stores the latest linked item batch.
	linked []int64
	// forumEnabled stores entitlement state.
	forumEnabled bool
	// removedPlayerID stores the latest removed member.
	removedPlayerID int64
	// returned stores the deterministic headquarters cleanup result.
	returned grouprecord.FurnitureReturn
}

// Membership returns the configured active membership.
func (store *membershipStore) Membership(_ context.Context, _ int64, playerID int64) (grouprecord.Membership, bool, error) {
	member, found := store.members[playerID]
	return member, found, nil
}

// Pending reports the configured deterministic membership request.
func (store *membershipStore) Pending(_ context.Context, _ int64, playerID int64) (bool, error) {
	return store.pending[playerID], nil
}

// Group returns the configured active group.
func (store *membershipStore) Group(context.Context, int64, bool) (grouprecord.Group, bool, error) {
	return store.group, store.group.ID > 0, nil
}

// BadgeParts returns no layers for the deterministic fixture.
func (store *membershipStore) BadgeParts(context.Context, int64) ([]grouprecord.BadgePart, error) {
	return nil, nil
}

// PlayerGroups returns the configured player projection.
func (store *membershipStore) PlayerGroups(context.Context, int64) ([]grouprecord.PlayerGroup, error) {
	return store.playerGroups, nil
}

// MemberPage records and returns one deterministic roster page.
func (store *membershipStore) MemberPage(_ context.Context, _ int64, page int32, pageSize int32, query string, level int32) (grouprecord.MemberPage, error) {
	store.memberPageLevel = level
	return grouprecord.MemberPage{Group: store.group, Page: page, PageSize: pageSize, Level: level, Query: query}, nil
}

// Join inserts one deterministic active membership.
func (store *membershipStore) Join(_ context.Context, groupID int64, playerID int64, _ int, _ int, _ int) (grouprecord.Membership, bool, bool, error) {
	if member, found := store.members[playerID]; found {
		return member, false, false, nil
	}
	member := grouprecord.Membership{GroupID: groupID, PlayerID: playerID, Role: grouprecord.Member}
	store.members[playerID] = member
	store.group.MemberCount++
	store.playerGroups = []grouprecord.PlayerGroup{{Group: store.group, Role: member.Role}}
	return member, false, true, nil
}

// AcceptRequest converts one deterministic request into a member.
func (store *membershipStore) AcceptRequest(_ context.Context, groupID int64, playerID int64, _ int) (grouprecord.Membership, error) {
	member := grouprecord.Membership{GroupID: groupID, PlayerID: playerID, Role: grouprecord.Member}
	delete(store.pending, playerID)
	store.members[playerID] = member
	store.group.MemberCount++
	store.group.PendingCount--
	return member, nil
}

// ChangeRole changes one configured role.
func (store *membershipStore) ChangeRole(_ context.Context, groupID int64, playerID int64, role grouprecord.Role) (grouprecord.Membership, error) {
	member := grouprecord.Membership{GroupID: groupID, PlayerID: playerID, Role: role}
	store.members[playerID] = member
	return member, nil
}

// RemoveMember removes one configured non-owner membership.
func (store *membershipStore) RemoveMember(_ context.Context, _ int64, playerID int64, _ int) (grouprecord.FurnitureReturn, error) {
	member, found := store.members[playerID]
	if !found {
		return grouprecord.FurnitureReturn{}, nil
	}
	if member.Role == grouprecord.Owner {
		return grouprecord.FurnitureReturn{}, grouprecord.ErrForbidden
	}
	delete(store.members, playerID)
	store.removedPlayerID = playerID
	store.playerGroups = nil
	store.group.MemberCount--
	return store.returned, nil
}

// SetFavorite records one validated favorite mutation.
func (store *membershipStore) SetFavorite(_ context.Context, _ int64, groupID *int64) error {
	store.favorite = groupID
	for index := range store.playerGroups {
		store.playerGroups[index].Favorite = groupID != nil && store.playerGroups[index].Group.ID == *groupID
	}
	return nil
}

// LinkFurniture records linked item identifiers.
func (store *membershipStore) LinkFurniture(_ context.Context, _ int64, itemIDs []int64) error {
	store.linked = append([]int64(nil), itemIDs...)
	return nil
}

// EnableForum activates the deterministic entitlement once.
func (store *membershipStore) EnableForum(context.Context, int64) (bool, error) {
	if store.forumEnabled {
		return false, nil
	}
	store.forumEnabled = true
	return true, nil
}

// TestChangeRoleRequiresOwner verifies the social role boundary.
func TestChangeRoleRequiresOwner(t *testing.T) {
	store := &membershipStore{group: grouprecord.Group{ID: 3}, members: map[int64]grouprecord.Membership{1: {GroupID: 3, PlayerID: 1, Role: grouprecord.Admin}, 2: {GroupID: 3, PlayerID: 2, Role: grouprecord.Member}}}
	service := New(groupconfig.Config{}, store, nil, groupruntime.NewCache(), nil, nil, nil)
	if _, err := service.ChangeRole(context.Background(), 1, 3, 2, grouprecord.Admin); !errors.Is(err, grouprecord.ErrForbidden) {
		t.Fatalf("expected forbidden, got %v", err)
	}
	store.members[1] = grouprecord.Membership{GroupID: 3, PlayerID: 1, Role: grouprecord.Owner}
	member, err := service.ChangeRole(context.Background(), 1, 3, 2, grouprecord.Admin)
	if err != nil || member.Role != grouprecord.Admin {
		t.Fatalf("member=%#v err=%v", member, err)
	}
}

// TestValidateCatalogEnforcesForumOwner verifies product-specific membership policy.
func TestValidateCatalogEnforcesForumOwner(t *testing.T) {
	store := &membershipStore{group: grouprecord.Group{ID: 3}, members: map[int64]grouprecord.Membership{1: {GroupID: 3, PlayerID: 1, Role: grouprecord.Member}}}
	service := New(groupconfig.Config{}, store, nil, groupruntime.NewCache(), nil, nil, nil)
	if err := service.ValidateCatalog(context.Background(), 1, 3, false); err != nil {
		t.Fatal(err)
	}
	if err := service.ValidateCatalog(context.Background(), 1, 3, true); !errors.Is(err, grouprecord.ErrForbidden) {
		t.Fatalf("expected forbidden forum product, got %v", err)
	}
	store.members[1] = grouprecord.Membership{GroupID: 3, PlayerID: 1, Role: grouprecord.Owner}
	if err := service.CommitCatalog(context.Background(), 1, 3, true, []int64{7, 8}); err != nil || !store.forumEnabled || len(store.linked) != 2 {
		t.Fatalf("enabled=%v linked=%v err=%v", store.forumEnabled, store.linked, err)
	}
}

// TestSetFavoriteRefreshesImmutableProjection verifies committed favorite refresh.
func TestSetFavoriteRefreshesImmutableProjection(t *testing.T) {
	group := grouprecord.Group{ID: 3}
	store := &membershipStore{group: group, members: map[int64]grouprecord.Membership{}, playerGroups: []grouprecord.PlayerGroup{{Group: group, Role: grouprecord.Member, Favorite: true}}}
	cache := groupruntime.NewCache()
	service := New(groupconfig.Config{}, store, nil, cache, nil, nil, nil)
	if err := service.SetFavorite(context.Background(), 1, &group.ID); err != nil {
		t.Fatal(err)
	}
	snapshot, found := cache.Player(1)
	if !found || snapshot.FavoriteID != group.ID || store.favorite == nil || *store.favorite != group.ID {
		t.Fatalf("snapshot=%#v favorite=%v", snapshot, store.favorite)
	}
}

// TestJoinPublishesOnlyForCommittedChanges verifies duplicate clicks stay side-effect free.
func TestJoinPublishesOnlyForCommittedChanges(t *testing.T) {
	store := &membershipStore{group: grouprecord.Group{ID: 3}, members: make(map[int64]grouprecord.Membership)}
	events := bus.New()
	published := 0
	_, err := events.Subscribe(changedevent.Name, 0, func(context.Context, bus.Event) error {
		published++
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	service := New(groupconfig.Config{}, store, nil, groupruntime.NewCache(), nil, nil, events)
	if _, _, err = service.Join(context.Background(), 7, 3); err != nil {
		t.Fatal(err)
	}
	if _, _, err = service.Join(context.Background(), 7, 3); err != nil || published != 1 {
		t.Fatalf("published=%d err=%v", published, err)
	}
}

// TestRemoveSelfRefreshesMembershipInformation verifies the room widget can project a successful leave.
func TestRemoveSelfRefreshesMembershipInformation(t *testing.T) {
	group := grouprecord.Group{ID: 3, MemberCount: 2}
	store := &membershipStore{
		group:        group,
		members:      map[int64]grouprecord.Membership{7: {GroupID: group.ID, PlayerID: 7, Role: grouprecord.Member}},
		playerGroups: []grouprecord.PlayerGroup{{Group: group, Role: grouprecord.Member, Favorite: true}},
		returned:     grouprecord.FurnitureReturn{RoomID: 131, Items: []grouprecord.ReturnedFurniture{{ItemID: 39, OwnerPlayerID: 7}}},
	}
	service := New(groupconfig.Config{}, store, nil, groupruntime.NewCache(), nil, nil, nil)
	returned, err := service.Remove(context.Background(), 7, group.ID, 7)
	if err != nil {
		t.Fatal(err)
	}
	updated, _, member, pending, favorite, err := service.Information(context.Background(), 7, group.ID)
	if err != nil || member || pending || favorite || updated.MemberCount != 1 || store.removedPlayerID != 7 || returned != 1 {
		t.Fatalf("group=%#v member=%t pending=%t favorite=%t removed=%d returned=%d err=%v", updated, member, pending, favorite, store.removedPlayerID, returned, err)
	}
}
