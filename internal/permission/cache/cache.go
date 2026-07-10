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
)

// entry stores one local serialized cache value.
type entry struct {
	// value stores serialized cache data.
	value []byte
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
	// local stores serialized process-local entries.
	local map[string]entry
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

	return &Cache{redis: redis, log: log, localTTL: localTTL, sharedTTL: sharedTTL, local: make(map[string]entry)}
}

// load reads one typed cache value or invokes its persistence loader.
func load[T any](ctx context.Context, cache *Cache, key string, loader func(context.Context) (T, error)) (T, error) {
	var value T
	data, found := cache.find(ctx, key)
	if found {
		if err := json.Unmarshal(data, &value); err == nil {
			return value, nil
		}
		cache.removeLocal(key)
	}

	value, err := loader(ctx)
	if err != nil {
		return value, err
	}
	data, err = json.Marshal(value)
	if err != nil {
		return value, fmt.Errorf("encode permission cache %s: %w", key, err)
	}
	cache.store(ctx, key, data)

	return value, nil
}

// groupValue stores an optional permission group cache result.
type groupValue struct {
	// Group stores the loaded permission group.
	Group permissionmodel.Group `json:"group"`
	// Found reports whether the group exists.
	Found bool `json:"found"`
}
