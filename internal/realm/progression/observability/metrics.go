// Package observability owns lock-free progression counters and gauges.
package observability

import (
	"sync"
	"sync/atomic"
)

// Snapshot stores one consistent-enough administrative telemetry view.
type Snapshot struct {
	// Triggers stores processed trigger counts by bounded catalog key.
	Triggers map[string]uint64 `json:"triggers"`
	// LevelUps stores paid achievement level crossings.
	LevelUps uint64 `json:"levelUps"`
	// Rewards stores paid reward counts by bounded reward kind.
	Rewards map[string]uint64 `json:"rewards"`
	// QuestsCompleted stores completed quest transitions.
	QuestsCompleted uint64 `json:"questsCompleted"`
	// QueueDepth stores queued triggers awaiting aggregation.
	QueueDepth int64 `json:"queueDepth"`
	// PendingFlushes stores aggregated trigger keys awaiting persistence.
	PendingFlushes int64 `json:"pendingFlushes"`
	// CacheDefinitions stores achievement definitions in the current generation.
	CacheDefinitions int64 `json:"cacheDefinitions"`
}

// Metrics stores lock-free progression telemetry.
type Metrics struct {
	// triggers stores counters allocated once per catalog trigger key.
	triggers sync.Map
	// rewards stores counters allocated once per supported reward kind.
	rewards sync.Map
	// levelUps counts achievement level crossings.
	levelUps atomic.Uint64
	// questsCompleted counts completed quests.
	questsCompleted atomic.Uint64
	// queueDepth stores queued trigger work.
	queueDepth atomic.Int64
	// pendingFlushes stores aggregated pending keys.
	pendingFlushes atomic.Int64
	// cacheDefinitions stores the active catalog size.
	cacheDefinitions atomic.Int64
}

// New creates empty progression telemetry.
func New() *Metrics { return &Metrics{} }

// RecordTrigger increments one catalog-bounded trigger counter.
func (metrics *Metrics) RecordTrigger(key string) {
	if metrics == nil || key == "" {
		return
	}
	if counter, ok := metrics.triggers.Load(key); ok {
		counter.(*atomic.Uint64).Add(1)
		return
	}
	counter, _ := metrics.triggers.LoadOrStore(key, &atomic.Uint64{})
	counter.(*atomic.Uint64).Add(1)
}

// RecordLevelUps increments paid achievement level crossings.
func (metrics *Metrics) RecordLevelUps(count int) {
	if metrics != nil && count > 0 {
		metrics.levelUps.Add(uint64(count))
	}
}

// RecordReward increments one bounded reward-kind counter.
func (metrics *Metrics) RecordReward(kind string) {
	if metrics == nil || kind == "" {
		return
	}
	if counter, ok := metrics.rewards.Load(kind); ok {
		counter.(*atomic.Uint64).Add(1)
		return
	}
	counter, _ := metrics.rewards.LoadOrStore(kind, &atomic.Uint64{})
	counter.(*atomic.Uint64).Add(1)
}

// RecordQuestCompleted increments completed quests.
func (metrics *Metrics) RecordQuestCompleted() {
	if metrics != nil {
		metrics.questsCompleted.Add(1)
	}
}

// AddQueue adjusts the queued trigger gauge.
func (metrics *Metrics) AddQueue(delta int64) {
	if metrics != nil {
		metrics.queueDepth.Add(delta)
	}
}

// SetPending stores aggregated pending trigger keys.
func (metrics *Metrics) SetPending(count int) {
	if metrics != nil {
		metrics.pendingFlushes.Store(int64(count))
	}
}

// SetCacheDefinitions stores the active achievement definition count.
func (metrics *Metrics) SetCacheDefinitions(count int) {
	if metrics != nil {
		metrics.cacheDefinitions.Store(int64(count))
	}
}

// Snapshot returns one administrative telemetry view.
func (metrics *Metrics) Snapshot() Snapshot {
	value := Snapshot{Triggers: map[string]uint64{}, Rewards: map[string]uint64{}}
	if metrics == nil {
		return value
	}
	metrics.triggers.Range(func(key any, counter any) bool {
		value.Triggers[key.(string)] = counter.(*atomic.Uint64).Load()
		return true
	})
	metrics.rewards.Range(func(key any, counter any) bool {
		value.Rewards[key.(string)] = counter.(*atomic.Uint64).Load()
		return true
	})
	value.LevelUps = metrics.levelUps.Load()
	value.QuestsCompleted = metrics.questsCompleted.Load()
	value.QueueDepth = metrics.queueDepth.Load()
	value.PendingFlushes = metrics.pendingFlushes.Load()
	value.CacheDefinitions = metrics.cacheDefinitions.Load()
	return value
}
