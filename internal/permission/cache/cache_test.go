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
