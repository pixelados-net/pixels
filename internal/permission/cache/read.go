package cache

import (
	"context"
	"fmt"
	"time"

	permissionmodel "github.com/niflaot/pixels/internal/permission/model"
	"go.uber.org/zap"
)

// PlayerNodes returns cached direct player grants.
func (cache *Cache) PlayerNodes(ctx context.Context, playerID int64, loader func(context.Context) ([]permissionmodel.Grant, error)) ([]permissionmodel.Grant, error) {
	return load(ctx, cache, playerNodesKey(playerID), loader)
}

// PlayerGroups returns cached player group memberships.
func (cache *Cache) PlayerGroups(ctx context.Context, playerID int64, loader func(context.Context) ([]permissionmodel.Group, error)) ([]permissionmodel.Group, error) {
	return load(ctx, cache, playerGroupsKey(playerID), loader)
}

// Group returns one cached optional permission group.
func (cache *Cache) Group(ctx context.Context, groupID int64, loader func(context.Context) (permissionmodel.Group, bool, error)) (permissionmodel.Group, bool, error) {
	value, err := load(ctx, cache, groupKey(groupID), func(ctx context.Context) (groupValue, error) {
		group, found, loadErr := loader(ctx)
		return groupValue{Group: group, Found: found}, loadErr
	})

	return value.Group, value.Found, err
}

// GroupNodes returns cached group grants.
func (cache *Cache) GroupNodes(ctx context.Context, groupID int64, loader func(context.Context) ([]permissionmodel.Grant, error)) ([]permissionmodel.Grant, error) {
	return load(ctx, cache, groupNodesKey(groupID), loader)
}

// find reads one local or shared serialized value.
func (cache *Cache) find(ctx context.Context, key string) ([]byte, bool) {
	cache.mutex.RLock()
	local, found := cache.local[key]
	cache.mutex.RUnlock()
	if found && time.Now().Before(local.expiresAt) {
		return append([]byte{}, local.value...), true
	}
	if found {
		cache.removeLocal(key)
	}
	if cache.redis == nil {
		return nil, false
	}

	data, found, err := cache.redis.Find(ctx, key)
	if err != nil {
		cache.log.Warn("permission shared cache read failed", zap.String("key", key), zap.Error(err))
		return nil, false
	}
	if found {
		cache.storeLocal(key, data)
	}

	return data, found
}

// store writes one local and best-effort shared value.
func (cache *Cache) store(ctx context.Context, key string, value []byte) {
	cache.storeLocal(key, value)
	if cache.redis != nil {
		if err := cache.redis.Set(ctx, key, value, cache.sharedTTL); err != nil {
			cache.log.Warn("permission shared cache write failed", zap.String("key", key), zap.Error(err))
		}
	}
}

// storeLocal writes one process-local value.
func (cache *Cache) storeLocal(key string, value []byte) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	cache.local[key] = entry{value: append([]byte{}, value...), expiresAt: time.Now().Add(cache.localTTL)}
}

// playerNodesKey returns the direct player grants cache key.
func playerNodesKey(playerID int64) string {
	return fmt.Sprintf("permission:player:%d:nodes", playerID)
}

// playerGroupsKey returns the player memberships cache key.
func playerGroupsKey(playerID int64) string {
	return fmt.Sprintf("permission:player:%d:groups", playerID)
}

// groupKey returns the group record cache key.
func groupKey(groupID int64) string {
	return fmt.Sprintf("permission:group:%d", groupID)
}

// groupNodesKey returns the group grants cache key.
func groupNodesKey(groupID int64) string {
	return fmt.Sprintf("permission:group:%d:nodes", groupID)
}
