// Package cache contains shared and process-local permission caching.
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	permissionmodel "github.com/niflaot/pixels/internal/permission/model"
	redispkg "github.com/niflaot/pixels/pkg/redis"
	"go.uber.org/zap"
)

const (
	// DefaultLocalTTL stores the maximum process-local staleness.
	DefaultLocalTTL = 5 * time.Second
	// DefaultSharedTTL stores the Redis safety expiration.
	DefaultSharedTTL = 10 * time.Minute
	// localSweepInterval stores writes between expired-entry sweeps.
	localSweepInterval uint64 = 256
)

// entry stores one process-local cache value.
type entry struct {
	// value stores one typed process-local cache value.
	value any
	// expiresAt stores local expiration time.
	expiresAt time.Time
}

// Cache stores permission fragments locally and in Redis.
type Cache struct {
	// redis stores shared cache values when configured.
	redis *redispkg.Client
	// log records non-fatal shared cache failures.
	log *zap.Logger
	// localTTL controls local expiration.
	localTTL time.Duration
	// sharedTTL controls Redis expiration.
	sharedTTL time.Duration
	// mutex protects local entries.
	mutex sync.RWMutex
	// local stores typed process-local entries.
	local map[cacheKey]entry
	// writes counts local writes between expiration sweeps.
	writes uint64
}

// New creates a two-layer permission cache.
func New(redis *redispkg.Client, log *zap.Logger) *Cache {
	return NewWithTTL(redis, log, DefaultLocalTTL, DefaultSharedTTL)
}

// NewWithTTL creates a permission cache with explicit expiration durations.
func NewWithTTL(redis *redispkg.Client, log *zap.Logger, localTTL time.Duration, sharedTTL time.Duration) *Cache {
	if log == nil {
		log = zap.NewNop()
	}
	if localTTL <= 0 {
		localTTL = DefaultLocalTTL
	}
	if sharedTTL <= 0 {
		sharedTTL = DefaultSharedTTL
	}

	return &Cache{redis: redis, log: log, localTTL: localTTL, sharedTTL: sharedTTL, local: make(map[cacheKey]entry)}
}

// load reads one typed cache value or invokes its persistence loader.
func load[T any](ctx context.Context, cache *Cache, key cacheKey, loader func(context.Context) (T, error)) (T, error) {
	if local, found := findLocal[T](cache, key); found {
		return local, nil
	}

	return loadMiss(ctx, cache, key, loader)
}

// loadMiss reads shared state or invokes the persistence loader.
func loadMiss[T any](ctx context.Context, cache *Cache, key cacheKey, loader func(context.Context) (T, error)) (T, error) {
	var value T
	data, found := cache.findShared(ctx, key)
	if found && json.Unmarshal(data, &value) == nil {
		cache.storeLocal(key, value)
		return value, nil
	}

	value, err := loader(ctx)
	if err != nil {
		return value, err
	}
	if err := store(ctx, cache, key, value); err != nil {
		return value, err
	}

	return value, nil
}

// findLocal reads one typed process-local cache value.
func findLocal[T any](cache *Cache, key cacheKey) (T, bool) {
	var zero T
	cache.mutex.RLock()
	local, found := cache.local[key]
	cache.mutex.RUnlock()
	if !found {
		return zero, false
	}
	if !time.Now().Before(local.expiresAt) {
		cache.removeLocal(key)
		return zero, false
	}
	value, valid := local.value.(T)
	if !valid {
		cache.removeLocal(key)
		return zero, false
	}

	return value, true
}

// store writes one typed local value and optional serialized shared value.
func store[T any](ctx context.Context, cache *Cache, key cacheKey, value T) error {
	cache.storeLocal(key, value)
	if cache.redis == nil {
		return nil
	}
	data, err := json.Marshal(value)
	if err != nil {
		cache.removeLocal(key)
		return fmt.Errorf("encode permission cache %s: %w", key.shared(), err)
	}
	sharedKey := key.shared()
	if err := cache.redis.Set(ctx, sharedKey, data, cache.sharedTTL); err != nil {
		cache.log.Warn("permission shared cache write failed", zap.String("key", sharedKey), zap.Error(err))
	}

	return nil
}

// groupValue stores an optional permission group cache result.
type groupValue struct {
	// Group stores the loaded permission group.
	Group permissionmodel.Group `json:"group"`
	// Found reports whether the group exists.
	Found bool `json:"found"`
}
