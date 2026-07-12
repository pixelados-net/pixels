package core

import (
	"context"
	"testing"
)

// TestIgnoredUsersPersistAndProject verifies directional ignore lifecycle.
func TestIgnoredUsersPersistAndProject(t *testing.T) {
	service, store := serviceFixture(t)
	player := addLivePlayer(t, service.live, 1, "demo", 0)
	item, added, err := service.IgnoreByName(context.Background(), 1, "alice")
	if err != nil || !added || item.PlayerID != 2 || !player.IsIgnoring(2) {
		t.Fatalf("ignore item=%#v added=%t err=%v", item, added, err)
	}
	items, err := service.Ignored(context.Background(), 1)
	if err != nil || len(items) != 1 || !store.ignored[[2]int64{1, 2}] {
		t.Fatalf("ignored items=%#v err=%v", items, err)
	}
	_, removed, err := service.UnignoreByName(context.Background(), 1, "alice")
	if err != nil || !removed || player.IsIgnoring(2) {
		t.Fatalf("unignore removed=%t err=%v", removed, err)
	}
}

// TestRelationshipViewersReturnsOnlyMatchingLiveProfiles verifies targeted observer lookup.
func TestRelationshipViewersReturnsOnlyMatchingLiveProfiles(t *testing.T) {
	service, _ := serviceFixture(t)
	first := addLivePlayer(t, service.live, 1, "demo", 0)
	second := addLivePlayer(t, service.live, 2, "alice", 0)
	first.ViewProfile(9)
	second.ViewProfile(8)
	viewers := service.RelationshipViewers(9)
	if len(viewers) != 1 || viewers[0] != 1 {
		t.Fatalf("unexpected viewers=%v", viewers)
	}
}
