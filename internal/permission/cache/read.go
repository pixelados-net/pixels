package cache

import (
	"context"
	"strconv"
	"time"

	permissionmodel "github.com/niflaot/pixels/internal/permission/model"
	"go.uber.org/zap"
)

// keyKind identifies one permission cache fragment family.
type keyKind uint8

const (
	// playerNodesKind identifies direct player grants.
	playerNodesKind keyKind = iota + 1
	// playerGroupsKind identifies player group memberships.
	playerGroupsKind
	// groupKind identifies one permission group record.
	groupKind
	// groupNodesKind identifies one group's grants.
	groupNodesKind
)

// cacheKey identifies one local permission cache fragment without allocation.
type cacheKey struct {
	// kind identifies the fragment family.
	kind keyKind
	// id identifies the player or group.
	id int64
}

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

// findShared reads one shared serialized value.
func (cache *Cache) findShared(ctx context.Context, key cacheKey) ([]byte, bool) {
	if cache.redis == nil {
		return nil, false
	}

	sharedKey := key.shared()
	data, found, err := cache.redis.Find(ctx, sharedKey)
	if err != nil {
		cache.log.Warn("permission shared cache read failed", zap.String("key", sharedKey), zap.Error(err))
		return nil, false
	}

	return data, found
}

// storeLocal writes one process-local value.
func (cache *Cache) storeLocal(key cacheKey, value any) {
	cache.mutex.Lock()
	defer cache.mutex.Unlock()
	now := time.Now()
	cache.local[key] = entry{value: value, expiresAt: now.Add(cache.localTTL)}
	cache.writes++
	if cache.writes%localSweepInterval == 0 {
		cache.sweepExpired(now)
	}
}

// sweepExpired removes expired entries while the cache mutex is held.
func (cache *Cache) sweepExpired(now time.Time) {
	for key, local := range cache.local {
		if !now.Before(local.expiresAt) {
			delete(cache.local, key)
		}
	}
}

// playerNodesKey returns the direct player grants cache key.
func playerNodesKey(playerID int64) cacheKey {
	return cacheKey{kind: playerNodesKind, id: playerID}
}

// playerGroupsKey returns the player memberships cache key.
func playerGroupsKey(playerID int64) cacheKey {
	return cacheKey{kind: playerGroupsKind, id: playerID}
}

// groupKey returns the group record cache key.
func groupKey(groupID int64) cacheKey {
	return cacheKey{kind: groupKind, id: groupID}
}

// groupNodesKey returns the group grants cache key.
func groupNodesKey(groupID int64) cacheKey {
	return cacheKey{kind: groupNodesKind, id: groupID}
}

// shared returns the Redis representation of one cache key.
func (key cacheKey) shared() string {
	id := strconv.FormatInt(key.id, 10)
	switch key.kind {
	case playerNodesKind:
		return "permission:player:" + id + ":nodes"
	case playerGroupsKind:
		return "permission:player:" + id + ":groups"
	case groupKind:
		return "permission:group:" + id
	case groupNodesKind:
		return "permission:group:" + id + ":nodes"
	default:
		return "permission:unknown:" + id
	}
}
