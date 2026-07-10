package service

import (
	"context"
	"testing"

	"github.com/niflaot/pixels/internal/permission"
	permissionmodel "github.com/niflaot/pixels/internal/permission/model"
	"github.com/niflaot/pixels/pkg/bus"
)

// TestServiceInspectionReadsGroupsAndAffectedPlayers verifies administration reads.
func TestServiceInspectionReadsGroupsAndAffectedPlayers(t *testing.T) {
	store := newFakeStore()
	store.groups[1] = groupForTest(1, "member", 0, nil)
	store.memberships[7] = []int64{1}
	service := New(store, nil, nil, nil)
	groups, err := service.Groups(context.Background())
	if err != nil || len(groups) != 1 {
		t.Fatalf("unexpected groups=%#v err=%v", groups, err)
	}
	players, err := service.AffectedPlayerIDs(context.Background(), 1)
	if err != nil || len(players) != 1 || players[0] != 7 {
		t.Fatalf("unexpected players=%#v err=%v", players, err)
	}
	if _, err := service.AffectedPlayerIDs(context.Background(), 0); err != ErrInvalidGroupID {
		t.Fatalf("expected invalid group id, got %v", err)
	}
	if _, found, err := service.PrimaryGroup(context.Background(), 9); err != nil || found {
		t.Fatalf("expected no primary group, found=%v err=%v", found, err)
	}
	if _, err := service.EffectiveNodes(context.Background(), 0); err != ErrInvalidPlayerID {
		t.Fatalf("expected invalid effective player id, got %v", err)
	}
}

// BenchmarkHasPermissionDirect measures a warmed direct override lookup.
func BenchmarkHasPermissionDirect(b *testing.B) {
	service := newTestService(storeWithDirect(true), nil)
	benchmarkPermission(b, service, testAction, true)
}

// BenchmarkHasPermissionInherited measures a warmed three-level group lookup.
func BenchmarkHasPermissionInherited(b *testing.B) {
	service := newTestService(storeWithDeepInheritance(), nil)
	benchmarkPermission(b, service, testAction, true)
}

// BenchmarkHasPermissionNoMatch measures a warmed default-deny lookup.
func BenchmarkHasPermissionNoMatch(b *testing.B) {
	service := newTestService(storeWithNoMatch(), nil)
	benchmarkPermission(b, service, testAction, false)
}

// BenchmarkHasPermissionLongInheritance measures the guarded deep-tree fallback.
func BenchmarkHasPermissionLongInheritance(b *testing.B) {
	service := newTestService(storeWithLongInheritance(20, false), nil)
	benchmarkPermission(b, service, testAction, true)
}

// BenchmarkHasPermissionParallel measures concurrent warmed resolution.
func BenchmarkHasPermissionParallel(b *testing.B) {
	service := newTestService(storeWithDeepInheritance(), nil)
	if _, err := service.HasPermission(context.Background(), 7, testAction); err != nil {
		b.Fatalf("warm permission cache: %v", err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(parallel *testing.PB) {
		for parallel.Next() {
			allowed, err := service.HasPermission(context.Background(), 7, testAction)
			if err != nil || !allowed {
				b.Fatalf("unexpected allowed=%v err=%v", allowed, err)
			}
		}
	})
}

// BenchmarkEffectiveNodes measures the administration permission projection.
func BenchmarkEffectiveNodes(b *testing.B) {
	service := newTestService(storeWithDeepInheritance(), nil)
	ctx := context.Background()
	if _, err := service.EffectiveNodes(ctx, 7); err != nil {
		b.Fatalf("warm permission cache: %v", err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		nodes, err := service.EffectiveNodes(ctx, 7)
		if err != nil || len(nodes) == 0 {
			b.Fatalf("unexpected nodes=%#v err=%v", nodes, err)
		}
	}
}

// benchmarkPermission measures one warmed permission resolution shape.
func benchmarkPermission(b *testing.B, service *Service, node permission.Node, expected bool) {
	b.Helper()
	ctx := context.Background()
	if _, err := service.HasPermission(ctx, 7, node); err != nil {
		b.Fatalf("warm permission cache: %v", err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		allowed, err := service.HasPermission(ctx, 7, node)
		if err != nil || allowed != expected {
			b.Fatalf("unexpected allowed=%v err=%v", allowed, err)
		}
	}
}

// storeWithDirect creates a direct override fixture.
func storeWithDirect(allowed bool) *fakeStore {
	store := storeWithDeepInheritance()
	store.playerNodes[7] = []permissionmodel.Grant{{Node: testAction, Allowed: allowed}}
	return store
}

// storeWithWeightedGroups creates conflicting group priority fixtures.
func storeWithWeightedGroups() *fakeStore {
	store := newFakeStore()
	store.groups[1] = groupForTest(1, "admin", 100, nil)
	store.groups[2] = groupForTest(2, "specific", 10, nil)
	store.memberships[7] = []int64{2, 1}
	store.groupNodes[1] = []permissionmodel.Grant{{Node: "permission.*", Allowed: false}}
	store.groupNodes[2] = []permissionmodel.Grant{{Node: testAction, Allowed: true}}
	return store
}

// storeWithYieldingGroup creates a high-priority unrelated grant fixture.
func storeWithYieldingGroup() *fakeStore {
	store := storeWithWeightedGroups()
	store.groupNodes[1] = []permissionmodel.Grant{{Node: testOther, Allowed: false}}
	return store
}

// storeWithSpecificParent creates parent specificity fixtures.
func storeWithSpecificParent() *fakeStore {
	store := newFakeStore()
	parentID := int64(1)
	store.groups[1] = groupForTest(1, "parent", 0, nil)
	store.groups[2] = groupForTest(2, "child", 10, &parentID)
	store.memberships[7] = []int64{2}
	store.groupNodes[1] = []permissionmodel.Grant{{Node: testAction, Allowed: false}}
	store.groupNodes[2] = []permissionmodel.Grant{{Node: "permission.test.*", Allowed: true}}
	return store
}

// storeWithChildOverride creates equal-specificity child override fixtures.
func storeWithChildOverride() *fakeStore {
	store := storeWithSpecificParent()
	store.groupNodes[2] = []permissionmodel.Grant{{Node: testAction, Allowed: true}}
	return store
}

// storeWithDeepInheritance creates a three-level inherited grant fixture.
func storeWithDeepInheritance() *fakeStore {
	store := newFakeStore()
	first := int64(1)
	second := int64(2)
	store.groups[1] = groupForTest(1, "member", 0, nil)
	store.groups[2] = groupForTest(2, "moderator", 50, &first)
	store.groups[3] = groupForTest(3, "admin", 100, &second)
	store.memberships[7] = []int64{3}
	store.groupNodes[1] = []permissionmodel.Grant{{Node: testAction, Allowed: true}}
	return store
}

// storeWithNoMatch creates unrelated grants.
func storeWithNoMatch() *fakeStore {
	store := storeWithDeepInheritance()
	store.groupNodes[1] = []permissionmodel.Grant{{Node: testOther, Allowed: true}}
	return store
}

// storeWithCycle creates cyclic inheritance fixtures.
func storeWithCycle() *fakeStore {
	store := newFakeStore()
	first := int64(1)
	second := int64(2)
	store.groups[1] = groupForTest(1, "first", 10, &second)
	store.groups[2] = groupForTest(2, "second", 5, &first)
	store.memberships[7] = []int64{1}
	return store
}

// fakePublisher records permission events.
type fakePublisher struct {
	// events stores published permission events.
	events []bus.Event
}

// Publish records one permission event.
func (publisher *fakePublisher) Publish(_ context.Context, event bus.Event) error {
	publisher.events = append(publisher.events, event)
	return nil
}

// upsertGrant creates or replaces one grant in a fixture collection.
func upsertGrant(grants []permissionmodel.Grant, node permission.Node, allowed bool) []permissionmodel.Grant {
	for index := range grants {
		if grants[index].Node == node {
			grants[index].Allowed = allowed
			return grants
		}
	}
	return append(grants, permissionmodel.Grant{Node: node, Allowed: allowed})
}

// deleteGrant removes one grant from a fixture collection.
func deleteGrant(grants []permissionmodel.Grant, node permission.Node) []permissionmodel.Grant {
	filtered := grants[:0]
	for _, grant := range grants {
		if grant.Node != node {
			filtered = append(filtered, grant)
		}
	}
	return filtered
}
