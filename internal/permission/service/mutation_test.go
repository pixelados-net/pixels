package service

import (
	"context"
	"errors"
	"testing"

	"github.com/niflaot/pixels/internal/permission"
	permissionmodel "github.com/niflaot/pixels/internal/permission/model"
)

// TestPermissionMutationsInvalidateAndResolve verifies mutations take effect immediately.
func TestPermissionMutationsInvalidateAndResolve(t *testing.T) {
	store := newFakeStore()
	store.groups[1] = groupForTest(1, "member", 0, nil)
	events := &fakePublisher{}
	service := newTestService(store, events)
	ctx := context.Background()

	if err := service.AddPlayerToGroup(ctx, 7, 1); err != nil {
		t.Fatalf("add membership: %v", err)
	}
	if err := service.GrantGroupNode(ctx, 1, testAction, true); err != nil {
		t.Fatalf("grant group node: %v", err)
	}
	allowed, err := service.HasPermission(ctx, 7, testAction)
	if err != nil || !allowed {
		t.Fatalf("expected group grant allowed=%v err=%v", allowed, err)
	}
	if err := service.GrantPlayerNode(ctx, 7, testAction, false); err != nil {
		t.Fatalf("grant player deny: %v", err)
	}
	allowed, err = service.HasPermission(ctx, 7, testAction)
	if err != nil || allowed {
		t.Fatalf("expected direct deny allowed=%v err=%v", allowed, err)
	}
	if err := service.RevokePlayerNode(ctx, 7, testAction); err != nil {
		t.Fatalf("revoke player node: %v", err)
	}
	if err := service.RevokeGroupNode(ctx, 1, testAction); err != nil {
		t.Fatalf("revoke group node: %v", err)
	}
	if err := service.RemovePlayerFromGroup(ctx, 7, 1); err != nil {
		t.Fatalf("remove membership: %v", err)
	}
	if len(events.events) != 6 {
		t.Fatalf("expected six permission events, got %d", len(events.events))
	}
}

// TestAssignDefaultGroupRequiresAndAssignsMember verifies new player membership.
func TestAssignDefaultGroupRequiresAndAssignsMember(t *testing.T) {
	store := newFakeStore()
	service := newTestService(store, nil)
	if err := service.AssignDefaultGroup(context.Background(), 7); !errors.Is(err, ErrGroupNotFound) {
		t.Fatalf("expected missing member group, got %v", err)
	}
	store.groups[1] = groupForTest(1, defaultGroupName, 0, nil)
	if err := service.AssignDefaultGroup(context.Background(), 7); err != nil {
		t.Fatalf("assign default group: %v", err)
	}
	if len(store.memberships[7]) != 1 || store.memberships[7][0] != 1 {
		t.Fatalf("unexpected memberships %#v", store.memberships[7])
	}
}

// TestGroupCreationUpdateAndCycleValidation verifies group mutation boundaries.
func TestGroupCreationUpdateAndCycleValidation(t *testing.T) {
	store := newFakeStore()
	store.groups[1] = groupForTest(1, "member", 0, nil)
	service := newTestService(store, nil)
	parentID := int64(1)

	created, err := service.CreateGroup(context.Background(), CreateGroupParams{Name: " moderator ", Weight: 50,
		BadgeURL: " https://cdn.example/moderator.png ", ParentGroupID: &parentID})
	if err != nil || created.Name != "moderator" || created.BadgeURL != "https://cdn.example/moderator.png" {
		t.Fatalf("unexpected created group %#v err=%v", created, err)
	}
	self := created.ID
	selfPointer := &self
	_, err = service.UpdateGroup(context.Background(), created.ID, UpdateGroupParams{ParentGroupID: &selfPointer})
	if err != ErrInheritanceCycle {
		t.Fatalf("expected cycle error, got %v", err)
	}
}

// TestGroupMutationRejectsInvalidBadgeURL verifies badge images use public web URLs.
func TestGroupMutationRejectsInvalidBadgeURL(t *testing.T) {
	service := newTestService(newFakeStore(), nil)
	_, err := service.CreateGroup(context.Background(), CreateGroupParams{Name: "moderator", BadgeURL: "file:///badge.png"})
	if !errors.Is(err, ErrInvalidGroup) {
		t.Fatalf("expected invalid group, got %v", err)
	}
}

// TestGroupMutationRejectsDuplicateNames verifies stable uniqueness conflicts.
func TestGroupMutationRejectsDuplicateNames(t *testing.T) {
	store := newFakeStore()
	store.groups[1] = groupForTest(1, "member", 0, nil)
	store.groups[2] = groupForTest(2, "admin", 100, nil)
	service := newTestService(store, nil)
	if _, err := service.CreateGroup(context.Background(), CreateGroupParams{Name: "member"}); !errors.Is(err, ErrConflict) {
		t.Fatalf("expected create conflict, got %v", err)
	}
	name := "member"
	if _, err := service.UpdateGroup(context.Background(), 2, UpdateGroupParams{Name: &name}); !errors.Is(err, ErrConflict) {
		t.Fatalf("expected update conflict, got %v", err)
	}
}

// TestGroupUpdateInvalidatesAffectedMembershipCache verifies immediate weight projection.
func TestGroupUpdateInvalidatesAffectedMembershipCache(t *testing.T) {
	store := newFakeStore()
	store.groups[1] = groupForTest(1, "member", 0, nil)
	store.memberships[7] = []int64{1}
	service := newTestService(store, nil)
	group, found, err := service.PrimaryGroup(context.Background(), 7)
	if err != nil || !found || group.Weight != 0 {
		t.Fatalf("unexpected initial group=%#v found=%v err=%v", group, found, err)
	}
	weight := int32(50)
	if _, err := service.UpdateGroup(context.Background(), 1, UpdateGroupParams{Weight: &weight}); err != nil {
		t.Fatalf("update group: %v", err)
	}
	group, found, err = service.PrimaryGroup(context.Background(), 7)
	if err != nil || !found || group.Weight != 50 {
		t.Fatalf("unexpected refreshed group=%#v found=%v err=%v", group, found, err)
	}
}

// TestMutationValidationRejectsUnknownInputs verifies ids and node catalog boundaries.
func TestMutationValidationRejectsUnknownInputs(t *testing.T) {
	service := newTestService(newFakeStore(), nil)
	cases := []error{
		service.GrantPlayerNode(context.Background(), 0, testAction, true),
		service.GrantPlayerNode(context.Background(), 7, permission.Node("unknown.node"), true),
		service.AddPlayerToGroup(context.Background(), 7, 0),
		service.GrantGroupNode(context.Background(), 99, testAction, true),
	}
	wanted := []error{ErrInvalidPlayerID, ErrInvalidNode, ErrInvalidGroupID, ErrGroupNotFound}
	for index, err := range cases {
		if err != wanted[index] {
			t.Fatalf("case %d: got %v want %v", index, err, wanted[index])
		}
	}
}

// CreateGroup creates one fixture group.
func (store *fakeStore) CreateGroup(_ context.Context, group permissionmodel.Group) (permissionmodel.Group, error) {
	group.ID = int64(len(store.groups) + 1)
	group.Version.Version = 1
	store.groups[group.ID] = group
	return group, nil
}

// UpdateGroup updates one fixture group.
func (store *fakeStore) UpdateGroup(_ context.Context, group permissionmodel.Group) (permissionmodel.Group, bool, error) {
	if _, found := store.groups[group.ID]; !found {
		return permissionmodel.Group{}, false, nil
	}
	group.Version.Version++
	store.groups[group.ID] = group
	return group, true, nil
}

// UpsertGroupNode creates or replaces one fixture group grant.
func (store *fakeStore) UpsertGroupNode(_ context.Context, groupID int64, node permission.Node, allowed bool) error {
	store.groupNodes[groupID] = upsertGrant(store.groupNodes[groupID], node, allowed)
	return nil
}

// DeleteGroupNode deletes one fixture group grant.
func (store *fakeStore) DeleteGroupNode(_ context.Context, groupID int64, node permission.Node) error {
	store.groupNodes[groupID] = deleteGrant(store.groupNodes[groupID], node)
	return nil
}

// AddPlayerToGroup adds one fixture membership.
func (store *fakeStore) AddPlayerToGroup(_ context.Context, playerID int64, groupID int64) error {
	for _, existing := range store.memberships[playerID] {
		if existing == groupID {
			return nil
		}
	}
	store.memberships[playerID] = append(store.memberships[playerID], groupID)
	return nil
}

// RemovePlayerFromGroup removes one fixture membership.
func (store *fakeStore) RemovePlayerFromGroup(_ context.Context, playerID int64, groupID int64) error {
	groups := store.memberships[playerID][:0]
	for _, existing := range store.memberships[playerID] {
		if existing != groupID {
			groups = append(groups, existing)
		}
	}
	store.memberships[playerID] = groups
	return nil
}

// UpsertPlayerNode creates or replaces one fixture direct grant.
func (store *fakeStore) UpsertPlayerNode(_ context.Context, playerID int64, node permission.Node, allowed bool) error {
	store.playerNodes[playerID] = upsertGrant(store.playerNodes[playerID], node, allowed)
	return nil
}

// DeletePlayerNode deletes one fixture direct grant.
func (store *fakeStore) DeletePlayerNode(_ context.Context, playerID int64, node permission.Node) error {
	store.playerNodes[playerID] = deleteGrant(store.playerNodes[playerID], node)
	return nil
}
