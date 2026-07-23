package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	permissionmodel "github.com/niflaot/pixels/internal/permission/model"
	redispkg "github.com/niflaot/pixels/pkg/redis"
	"go.uber.org/zap"
)

// TestCacheUsesLocalRedisAndInvalidation verifies both cache layers and invalidation.
func TestCacheUsesLocalRedisAndInvalidation(t *testing.T) {
	server := miniredis.RunT(t)
	redis := redispkg.New(redispkg.Config{Address: server.Addr()})
	defer func() {
		if err := redis.Close(); err != nil {
			t.Fatalf("close redis: %v", err)
		}
	}()
	cache := NewWithTTL(redis, zap.NewNop(), time.Minute, time.Hour)
	loads := 0
	loader := func(context.Context) ([]permissionmodel.Grant, error) {
		loads++
		return []permissionmodel.Grant{{Node: "cache.test", Allowed: true}}, nil
	}

	for range 2 {
		if _, err := cache.PlayerNodes(context.Background(), 7, loader); err != nil {
			t.Fatalf("load player nodes: %v", err)
		}
	}
	if loads != 1 {
		t.Fatalf("expected one persistence load, got %d", loads)
	}
	cache.removeLocal(playerNodesKey(7))
	if _, err := cache.PlayerNodes(context.Background(), 7, loader); err != nil || loads != 1 {
		t.Fatalf("expected Redis hit loads=%d err=%v", loads, err)
	}
	cache.InvalidatePlayerNodes(context.Background(), 7)
	if _, err := cache.PlayerNodes(context.Background(), 7, loader); err != nil || loads != 2 {
		t.Fatalf("expected reload after invalidation loads=%d err=%v", loads, err)
	}
}

// TestCacheFallsBackWithoutRedis verifies local operation and expiration.
func TestCacheFallsBackWithoutRedis(t *testing.T) {
	cache := NewWithTTL(nil, zap.NewNop(), time.Millisecond, time.Minute)
	loads := 0
	loader := func(context.Context) ([]permissionmodel.Group, error) {
		loads++
		return []permissionmodel.Group{{Name: "member"}}, nil
	}
	if _, err := cache.PlayerGroups(context.Background(), 7, loader); err != nil {
		t.Fatalf("load local groups: %v", err)
	}
	time.Sleep(2 * time.Millisecond)
	if _, err := cache.PlayerGroups(context.Background(), 7, loader); err != nil || loads != 2 {
		t.Fatalf("expected local expiration loads=%d err=%v", loads, err)
	}
}

// TestCacheSweepsExpiredEntries verifies opportunistic retention cleanup.
func TestCacheSweepsExpiredEntries(t *testing.T) {
	cache := NewWithTTL(nil, zap.NewNop(), time.Nanosecond, time.Minute)
	expired := cacheKey{kind: playerNodesKind, id: 999}
	cache.storeLocal(expired, []permissionmodel.Grant{{Node: "cache.test", Allowed: true}})
	time.Sleep(time.Millisecond)
	for index := uint64(1); index < localSweepInterval; index++ {
		cache.storeLocal(cacheKey{kind: groupKind, id: int64(index)}, index)
	}
	cache.mutex.RLock()
	_, found := cache.local[expired]
	cache.mutex.RUnlock()
	if found {
		t.Fatal("expected expired entry to be swept")
	}
}

// TestCacheLoadsAndInvalidatesEveryFragment verifies complete fragment behavior.
func TestCacheLoadsAndInvalidatesEveryFragment(t *testing.T) {
	cache := New(nil, nil)
	ctx := context.Background()
	groupLoads := 0
	groupLoader := func(context.Context) (permissionmodel.Group, bool, error) {
		groupLoads++
		return permissionmodel.Group{Name: "admin"}, true, nil
	}
	for range 2 {
		group, found, err := cache.Group(ctx, 2, groupLoader)
		if err != nil || !found || group.Name != "admin" {
			t.Fatalf("unexpected group=%#v found=%v err=%v", group, found, err)
		}
	}
	nodeLoads := 0
	nodeLoader := func(context.Context) ([]permissionmodel.Grant, error) {
		nodeLoads++
		return []permissionmodel.Grant{{Node: "catalog.*", Allowed: true}}, nil
	}
	if _, err := cache.GroupNodes(ctx, 2, nodeLoader); err != nil {
		t.Fatalf("load group nodes: %v", err)
	}
	cache.InvalidateGroup(ctx, 2)
	if _, _, err := cache.Group(ctx, 2, groupLoader); err != nil {
		t.Fatalf("reload group: %v", err)
	}
	if _, err := cache.GroupNodes(ctx, 2, nodeLoader); err != nil {
		t.Fatalf("reload group nodes: %v", err)
	}
	groupsLoader := func(context.Context) ([]permissionmodel.Group, error) {
		return []permissionmodel.Group{{Name: "member"}}, nil
	}
	if _, err := cache.PlayerGroups(ctx, 7, groupsLoader); err != nil {
		t.Fatalf("load player groups: %v", err)
	}
	cache.InvalidatePlayerGroups(ctx, 7)
	if groupLoads != 2 || nodeLoads != 2 {
		t.Fatalf("unexpected loads groups=%d nodes=%d", groupLoads, nodeLoads)
	}
}

// TestCacheRecoversFromWrongLocalTypeAndMalformedSharedValue verifies safe misses.
func TestCacheRecoversFromWrongLocalTypeAndMalformedSharedValue(t *testing.T) {
	server := miniredis.RunT(t)
	redis := redispkg.New(redispkg.Config{Address: server.Addr()})
	defer func() { _ = redis.Close() }()
	cache := NewWithTTL(redis, zap.NewNop(), time.Minute, time.Hour)
	cache.storeLocal(playerNodesKey(7), "wrong")
	loads := 0
	loader := func(context.Context) ([]permissionmodel.Grant, error) {
		loads++
		return []permissionmodel.Grant{{Node: "cache.test", Allowed: true}}, nil
	}
	if _, err := cache.PlayerNodes(context.Background(), 7, loader); err != nil || loads != 1 {
		t.Fatalf("recover local type loads=%d err=%v", loads, err)
	}
	cache.removeLocal(playerNodesKey(7))
	if err := redis.Set(context.Background(), playerNodesKey(7).shared(), []byte("{"), time.Hour); err != nil {
		t.Fatalf("write malformed shared value: %v", err)
	}
	if _, err := cache.PlayerNodes(context.Background(), 7, loader); err != nil || loads != 2 {
		t.Fatalf("recover shared value loads=%d err=%v", loads, err)
	}
}

// TestCacheKeySharedRepresentation verifies every Redis key family.
func TestCacheKeySharedRepresentation(t *testing.T) {
	cases := map[cacheKey]string{
		playerNodesKey(7):  "permission:player:7:nodes",
		playerGroupsKey(7): "permission:player:7:groups",
		groupKey(2):        "permission:group:2",
		groupNodesKey(2):   "permission:group:2:nodes",
		{kind: 99, id: 1}:  "permission:unknown:1",
	}
	for key, expected := range cases {
		if key.shared() != expected {
			t.Fatalf("expected %q, got %q", expected, key.shared())
		}
	}
}

// BenchmarkPlayerNodesLocalHit measures the permission cache hot path.
func BenchmarkPlayerNodesLocalHit(b *testing.B) {
	cache := NewWithTTL(nil, zap.NewNop(), time.Minute, time.Hour)
	ctx := context.Background()
	loader := func(context.Context) ([]permissionmodel.Grant, error) {
		return []permissionmodel.Grant{{Node: "cache.test", Allowed: true}}, nil
	}
	if _, err := cache.PlayerNodes(ctx, 7, loader); err != nil {
		b.Fatalf("warm cache: %v", err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		grants, err := cache.PlayerNodes(ctx, 7, loader)
		if err != nil || len(grants) != 1 {
			b.Fatalf("unexpected grants=%#v err=%v", grants, err)
		}
	}
}

// BenchmarkPlayerNodesRedisRefill measures shared-cache local refill cost.
func BenchmarkPlayerNodesRedisRefill(b *testing.B) {
	server := miniredis.RunT(b)
	redis := redispkg.New(redispkg.Config{Address: server.Addr()})
	b.Cleanup(func() { _ = redis.Close() })
	cache := NewWithTTL(redis, zap.NewNop(), time.Minute, time.Hour)
	ctx := context.Background()
	loader := func(context.Context) ([]permissionmodel.Grant, error) {
		return []permissionmodel.Grant{{Node: "cache.test", Allowed: true}}, nil
	}
	if _, err := cache.PlayerNodes(ctx, 7, loader); err != nil {
		b.Fatalf("warm shared cache: %v", err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for b.Loop() {
		cache.removeLocal(playerNodesKey(7))
		if _, err := cache.PlayerNodes(ctx, 7, loader); err != nil {
			b.Fatalf("refill cache: %v", err)
		}
	}
}
