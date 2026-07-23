package runtime

import (
	"container/heap"
	"context"
	"errors"
	"math/rand/v2"
	"sync"
	"sync/atomic"
	"time"

	roomwired "github.com/niflaot/pixels/internal/realm/room/world/wired"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/condition"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/effect"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// Engine owns compiled generations and bounded room execution.
type Engine struct {
	// config stores normalized safety budgets.
	config roomwired.Config
	// store loads durable configurations.
	store record.Store
	// compiler creates immutable generations.
	compiler *configuration.Compiler
	// matcher matches trigger candidates.
	matcher trigger.Matcher
	// conditions evaluates stack gates.
	conditions condition.Evaluator
	// effects executes stack effects.
	effects *effect.Executor
	// views resolves condition facts.
	views ViewProvider
	// scheduler queues delayed room-owned effects.
	scheduler Scheduler
	// activator projects executed-box animation without generating furniture events.
	activator Activator
	// rooms stores state pointers by room id.
	rooms sync.Map
	// generationID allocates unique generation identifiers.
	generationID atomic.Uint64
	// eventID allocates event identifiers.
	eventID atomic.Uint64
	// metrics stores lock-free low-cardinality execution counters.
	metrics metrics
}

// New creates a WIRED runtime engine.
func New(config roomwired.Config, store record.Store, compiler *configuration.Compiler, effects *effect.Executor, views ViewProvider, scheduler Scheduler, activator Activator) *Engine {
	return &Engine{config: config.Normalize(), store: store, compiler: compiler, matcher: trigger.New(), conditions: condition.New(), effects: effects, views: views, scheduler: scheduler, activator: activator}
}

// Reload loads, compiles, and atomically replaces one room generation.
func (engine *Engine) Reload(ctx context.Context, roomID int64, now time.Time) error {
	if !engine.config.Enabled {
		engine.Close(roomID)
		return nil
	}
	started := time.Now()
	records, err := engine.store.LoadRoom(ctx, roomID)
	if err != nil {
		engine.metrics.compileFailures.Add(1)
		return err
	}
	generation, err := engine.compiler.Compile(roomID, engine.generationID.Add(1), records)
	if err != nil {
		engine.metrics.compileFailures.Add(1)
		return err
	}
	loaded := &state{generation: generation, resetAt: now, unseen: make(map[configuration.Point]int)}
	for _, node := range generation.Triggers {
		kind := eventKind(node.Descriptor.Key)
		if kind != 0 {
			loaded.byKind[kind] = append(loaded.byKind[kind], node)
		}
	}
	loaded.timers = buildTimers(generation, now)
	previous, replaced := engine.rooms.Swap(roomID, loaded)
	if replaced {
		engine.releaseDelayed(previous.(*state))
	}
	engine.metrics.compileCount.Add(1)
	engine.metrics.compileNanoseconds.Add(uint64(time.Since(started)))
	return nil
}

// Close releases one room generation and every mutable cursor.
func (engine *Engine) Close(roomID int64) {
	value, found := engine.rooms.LoadAndDelete(roomID)
	if found {
		engine.releaseDelayed(value.(*state))
	}
}

// Loaded reports whether one room has an active compiled generation.
func (engine *Engine) Loaded(roomID int64) bool {
	_, found := engine.rooms.Load(roomID)
	return found
}

// Process executes stacks matching one room event.
func (engine *Engine) Process(ctx context.Context, event trigger.Event, now time.Time) (Trace, error) {
	value, found := engine.rooms.Load(event.RoomID)
	if !found || !engine.config.Enabled {
		return Trace{}, nil
	}
	if event.ID == 0 {
		event.ID = engine.eventID.Add(1)
	}
	if int(event.Kind) < len(engine.metrics.events) {
		engine.metrics.events[event.Kind].Add(1)
	}
	loaded := value.(*state)
	loaded.mutex.Lock()
	defer loaded.mutex.Unlock()
	return engine.processLocked(ctx, loaded, event, now)
}

// Cycle executes due timers with no catch-up burst.
func (engine *Engine) Cycle(ctx context.Context, roomID int64, now time.Time) error {
	value, found := engine.rooms.Load(roomID)
	if !found || !engine.config.Enabled {
		return nil
	}
	loaded := value.(*state)
	for {
		loaded.mutex.Lock()
		if len(loaded.timers) == 0 || loaded.timers[0].deadline.After(now) {
			loaded.mutex.Unlock()
			return nil
		}
		timer := heap.Pop(&loaded.timers).(timerEntry)
		if timer.period > 0 {
			timer.deadline = now.Add(timer.period)
			heap.Push(&loaded.timers, timer)
		}
		event := trigger.Event{ID: engine.eventID.Add(1), Kind: timer.kind, RoomID: roomID, SourceItem: timer.node.ItemID}
		_, err := engine.processTriggerLocked(ctx, loaded, timer.node, event, now)
		loaded.mutex.Unlock()
		if err != nil {
			return err
		}
	}
}

// Traces returns the room's bounded trace ring in oldest-first order.
func (engine *Engine) Traces(roomID int64) []Trace {
	value, found := engine.rooms.Load(roomID)
	if !found {
		return nil
	}
	loaded := value.(*state)
	loaded.mutex.Lock()
	defer loaded.mutex.Unlock()
	result := make([]Trace, loaded.traceCount)
	start := (loaded.traceNext - loaded.traceCount + len(loaded.traces)) % len(loaded.traces)
	for index := range result {
		result[index] = loaded.traces[(start+index)%len(loaded.traces)]
	}
	return result
}

// ResetTimers resets timer origin and rebuilds every deadline.
func (engine *Engine) ResetTimers(roomID int64, now time.Time) bool {
	value, found := engine.rooms.Load(roomID)
	if !found {
		return false
	}
	loaded := value.(*state)
	loaded.mutex.Lock()
	loaded.resetAt = now
	loaded.timers = buildTimers(loaded.generation, now)
	loaded.mutex.Unlock()
	return true
}

// IsCurrent reports whether a scheduled generation remains active.
func (engine *Engine) IsCurrent(roomID int64, generationID uint64) bool {
	value, found := engine.rooms.Load(roomID)
	return found && value.(*state).generation.ID == generationID
}

// Matches reports whether an active generation has a matching trigger candidate.
func (engine *Engine) Matches(event trigger.Event) bool {
	value, found := engine.rooms.Load(event.RoomID)
	if !found || !engine.config.Enabled {
		return false
	}
	loaded := value.(*state)
	loaded.mutex.Lock()
	defer loaded.mutex.Unlock()
	for _, candidate := range loaded.byKind[event.Kind] {
		if engine.matcher.Match(candidate, event) {
			return true
		}
	}
	return false
}

// Conflicts returns trigger sprites incompatible with one effect's actor requirements.
func (engine *Engine) Conflicts(roomID int64, itemID int64) []int32 {
	value, found := engine.rooms.Load(roomID)
	if !found {
		return nil
	}
	loaded := value.(*state)
	loaded.mutex.Lock()
	defer loaded.mutex.Unlock()
	node := loaded.generation.Nodes[itemID]
	if node == nil {
		return nil
	}
	stack := loaded.generation.Stacks[node.Point]
	result := make([]int32, 0)
	for _, candidate := range stack.Triggers {
		if node.Descriptor.Actor != 0 && candidate.Descriptor.Actor == 0 {
			result = append(result, candidate.SpriteID)
		}
	}
	return result
}

// joined joins non-nil execution errors.
func joined(current error, next error) error { return errors.Join(current, next) }

// randomIndex returns a deterministic per-event selector.
func randomIndex(eventID uint64, point configuration.Point, maximum int) int {
	if maximum <= 1 {
		return 0
	}
	seed := uint64(point.X+1)<<32 | uint64(uint32(point.Y+1))
	random := rand.New(rand.NewPCG(eventID, seed))
	return random.IntN(maximum)
}
