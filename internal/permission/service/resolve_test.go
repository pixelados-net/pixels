package service

import (
	"context"
	"sort"
	"testing"

	"github.com/niflaot/pixels/internal/permission"
	permissioncache "github.com/niflaot/pixels/internal/permission/cache"
	permissionmodel "github.com/niflaot/pixels/internal/permission/model"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
	"go.uber.org/zap"
)

var (
	// testAction identifies the concrete resolution fixture node.
	testAction = permission.RegisterNode("permission.test.action", "TEST_ACTION")
	// testOther identifies an unrelated resolution fixture node.
	testOther = permission.RegisterNode("permission.test.other", "")
)

// TestHasPermissionResolutionOrder verifies direct, weight, inheritance, and specificity rules.
func TestHasPermissionResolutionOrder(t *testing.T) {
	cases := []struct {
		name   string
		store  *fakeStore
		allow  bool
		wanted error
	}{
		{name: "direct deny beats exact group allow", store: storeWithDirect(false)},
		{name: "higher weight wildcard deny beats lower exact allow", store: storeWithWeightedGroups()},
		{name: "higher weight without match yields to lower group", store: storeWithYieldingGroup(), allow: true},
		{name: "parent exact beats child wildcard", store: storeWithSpecificParent()},
		{name: "child exact beats parent exact", store: storeWithChildOverride(), allow: true},
		{name: "three level inheritance resolves ancestor", store: storeWithDeepInheritance(), allow: true},
		{name: "missing match denies", store: storeWithNoMatch()},
		{name: "cycle fails closed", store: storeWithCycle(), wanted: ErrInheritanceCycle},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			service := newTestService(testCase.store, nil)
			allowed, err := service.HasPermission(context.Background(), 7, testAction)
			if err != testCase.wanted || allowed != testCase.allow {
				t.Fatalf("allowed=%v err=%v, want allowed=%v err=%v", allowed, err, testCase.allow, testCase.wanted)
			}
		})
	}
}

// TestEffectiveNodesPerksAndPrimaryGroup verifies permission inspection projections.
func TestEffectiveNodesPerksAndPrimaryGroup(t *testing.T) {
	service := newTestService(storeWithDeepInheritance(), nil)
	group, found, err := service.PrimaryGroup(context.Background(), 7)
	if err != nil || !found || group.ID != 3 {
		t.Fatalf("unexpected primary group %#v found=%v err=%v", group, found, err)
	}
	nodes, err := service.EffectiveNodes(context.Background(), 7)
	if err != nil || !containsDecision(nodes, testAction, true) {
		t.Fatalf("unexpected nodes %#v err=%v", nodes, err)
	}
	perks, err := service.EffectivePerks(context.Background(), 7)
	if err != nil || !containsString(perks, "TEST_ACTION") {
		t.Fatalf("unexpected perks %#v err=%v", perks, err)
	}
}

// TestHasPermissionValidatesInput verifies resolver identity and query boundaries.
func TestHasPermissionValidatesInput(t *testing.T) {
	service := newTestService(newFakeStore(), nil)
	if _, err := service.HasPermission(context.Background(), 0, testAction); err != ErrInvalidPlayerID {
		t.Fatalf("expected invalid player, got %v", err)
	}
	if _, err := service.HasPermission(context.Background(), 7, "permission.test.*"); err != ErrInvalidNode {
		t.Fatalf("expected invalid query node, got %v", err)
	}
}

// fakeStore stores permission service fixtures.
type fakeStore struct {
	// groups stores groups by id.
	groups map[int64]permissionmodel.Group
	// memberships stores player group ids.
	memberships map[int64][]int64
	// groupNodes stores grants by group id.
	groupNodes map[int64][]permissionmodel.Grant
	// playerNodes stores direct grants by player id.
	playerNodes map[int64][]permissionmodel.Grant
}

// newFakeStore creates empty permission fixtures.
func newFakeStore() *fakeStore {
	return &fakeStore{groups: make(map[int64]permissionmodel.Group), memberships: make(map[int64][]int64),
		groupNodes: make(map[int64][]permissionmodel.Grant), playerNodes: make(map[int64][]permissionmodel.Grant)}
}

// ListGroups lists fixture groups.
func (store *fakeStore) ListGroups(context.Context) ([]permissionmodel.Group, error) {
	groups := make([]permissionmodel.Group, 0, len(store.groups))
	for _, group := range store.groups {
		groups = append(groups, group)
	}
	sort.Slice(groups, func(left int, right int) bool { return groups[left].Weight > groups[right].Weight })
	return groups, nil
}

// ListGroupsByPlayer lists fixture memberships.
func (store *fakeStore) ListGroupsByPlayer(_ context.Context, playerID int64) ([]permissionmodel.Group, error) {
	groups := make([]permissionmodel.Group, 0, len(store.memberships[playerID]))
	for _, groupID := range store.memberships[playerID] {
		groups = append(groups, store.groups[groupID])
	}
	return groups, nil
}

// FindGroupByID finds one fixture group.
func (store *fakeStore) FindGroupByID(_ context.Context, groupID int64) (permissionmodel.Group, bool, error) {
	group, found := store.groups[groupID]
	return group, found, nil
}

// FindGroupByName finds one fixture group by name.
func (store *fakeStore) FindGroupByName(_ context.Context, name string) (permissionmodel.Group, bool, error) {
	for _, group := range store.groups {
		if group.Name == name {
			return group, true, nil
		}
	}
	return permissionmodel.Group{}, false, nil
}

// ListGroupNodes lists fixture group grants.
func (store *fakeStore) ListGroupNodes(_ context.Context, groupID int64) ([]permissionmodel.Grant, error) {
	return append([]permissionmodel.Grant{}, store.groupNodes[groupID]...), nil
}

// ListPlayerNodes lists fixture direct grants.
func (store *fakeStore) ListPlayerNodes(_ context.Context, playerID int64) ([]permissionmodel.Grant, error) {
	return append([]permissionmodel.Grant{}, store.playerNodes[playerID]...), nil
}

// ListAffectedPlayerIDs lists fixture members of one group.
func (store *fakeStore) ListAffectedPlayerIDs(_ context.Context, groupID int64) ([]int64, error) {
	players := make([]int64, 0)
	for playerID, groups := range store.memberships {
		for _, candidate := range groups {
			if candidate == groupID {
				players = append(players, playerID)
			}
		}
	}
	return players, nil
}

// newTestService creates permission behavior without shared Redis state.
func newTestService(store *fakeStore, events *fakePublisher) *Service {
	if events == nil {
		return New(store, permissioncache.New(nil, zap.NewNop()), nil, zap.NewNop())
	}
	return New(store, permissioncache.New(nil, zap.NewNop()), events, zap.NewNop())
}

// groupForTest creates one fixture group.
func groupForTest(id int64, name string, weight int32, parent *int64) permissionmodel.Group {
	return permissionmodel.Group{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: id}, Version: sharedmodel.Version{Version: 1}},
		Name: name, Weight: weight, ParentGroupID: parent}
}

// containsDecision reports whether a resolved node list contains a decision.
func containsDecision(nodes []ResolvedNode, node permission.Node, allowed bool) bool {
	for _, resolved := range nodes {
		if resolved.Node == node && resolved.Allowed == allowed {
			return true
		}
	}
	return false
}

// containsString reports whether a string collection contains one value.
func containsString(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}
