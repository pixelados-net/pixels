package cache

import (
	"context"

	"go.uber.org/zap"
)

// InvalidatePlayerNodes invalidates one player's direct grants.
func (cache *Cache) InvalidatePlayerNodes(ctx context.Context, playerID int64) {
	cache.delete(ctx, playerNodesKey(playerID))
}

// InvalidatePlayerGroups invalidates one player's group memberships.
func (cache *Cache) InvalidatePlayerGroups(ctx context.Context, playerID int64) {
	cache.delete(ctx, playerGroupsKey(playerID))
}

// InvalidateGroup invalidates one group record and its grants.
func (cache *Cache) InvalidateGroup(ctx context.Context, groupID int64) {
	cache.delete(ctx, groupKey(groupID))
	cache.delete(ctx, groupNodesKey(groupID))
}

// delete removes one local and shared key.
func (cache *Cache) delete(ctx context.Context, key cacheKey) {
	cache.removeLocal(key)
	if cache.redis != nil {
		sharedKey := key.shared()
		if err := cache.redis.Delete(ctx, sharedKey); err != nil {
			cache.log.Warn("permission shared cache invalidation failed", zap.String("key", sharedKey), zap.Error(err))
		}
	}
}

// removeLocal removes one process-local key.
func (cache *Cache) removeLocal(key cacheKey) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	delete(cache.local, key)
}
