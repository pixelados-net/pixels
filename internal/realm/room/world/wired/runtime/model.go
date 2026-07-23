// Package runtime executes immutable WIRED room generations.
package runtime

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/condition"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/effect"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// Scheduler queues delayed work on the owning room lifecycle.
type Scheduler interface {
	// Schedule queues work only while the room and generation remain active.
	Schedule(int64, uint64, time.Duration, func(time.Time)) bool
}

// ViewProvider resolves read-only condition facts for one active room.
type ViewProvider interface {
	// View returns the room condition view.
	View(int64) (condition.View, bool)
}

// Activator projects WIRED box execution without emitting furniture behavior events.
type Activator interface {
	// Activate toggles one executed WIRED item through its compact visual state.
	Activate(context.Context, int64, int64) error
}

// Trace stores a bounded execution summary.
type Trace struct {
	// ID identifies the source event.
	ID uint64
	// Kind classifies the source event.
	Kind trigger.Kind
	// Stacks stores visited stack count.
	Stacks int
	// Effects stores attempted effect count.
	Effects int
	// BudgetExhausted reports whether safety budgets stopped execution.
	BudgetExhausted bool
	// StartedAt stores trace start time.
	StartedAt time.Time
	// Duration stores trace execution duration.
	Duration time.Duration
}

// MetricsSnapshot contains low-cardinality WIRED runtime counters.
type MetricsSnapshot struct {
	// Events stores processed events indexed by trigger kind.
	Events [18]uint64 `json:"events"`
	// StackResults stores passed, failed, and errored stack evaluations.
	StackResults [3]uint64 `json:"stackResults"`
	// EffectResults stores applied, skipped, and blocked effect results by status value.
	EffectResults [4]uint64 `json:"effectResults"`
	// BudgetExhausted stores traces stopped by a configured safety budget.
	BudgetExhausted uint64 `json:"budgetExhausted"`
	// CompileFailures stores failed room generation loads or compilations.
	CompileFailures uint64 `json:"compileFailures"`
	// DelayedTasks stores currently outstanding delayed effects.
	DelayedTasks int64 `json:"delayedTasks"`
	// CompileCount stores successful room generation compilations.
	CompileCount uint64 `json:"compileCount"`
	// CompileNanoseconds stores cumulative successful compilation duration.
	CompileNanoseconds uint64 `json:"compileNanoseconds"`
	// TraceCount stores completed execution traces.
	TraceCount uint64 `json:"traceCount"`
	// TraceNanoseconds stores cumulative trace duration.
	TraceNanoseconds uint64 `json:"traceNanoseconds"`
}

// metrics stores lock-free low-cardinality execution counters.
type metrics struct {
	// events stores counters indexed by trigger kind.
	events [18]atomic.Uint64
	// stackResults stores pass, fail, and error counters.
	stackResults [3]atomic.Uint64
	// effectResults stores counters indexed by effect status.
	effectResults [4]atomic.Uint64
	// budgetExhausted counts traces stopped by budgets.
	budgetExhausted atomic.Uint64
	// compileFailures counts failed generation loads.
	compileFailures atomic.Uint64
	// delayedTasks gauges outstanding delayed effects.
	delayedTasks atomic.Int64
	// compileCount counts successful generation loads.
	compileCount atomic.Uint64
	// compileNanoseconds accumulates compilation duration.
	compileNanoseconds atomic.Uint64
	// traceCount counts completed traces.
	traceCount atomic.Uint64
	// traceNanoseconds accumulates trace duration.
	traceNanoseconds atomic.Uint64
}

// state stores room-owned mutable cursors around an immutable generation.
type state struct {
	// mutex serializes room events, timer cursors, and trace state.
	mutex sync.Mutex
	// generation stores immutable compiled room nodes.
	generation *configuration.Generation
	// byKind indexes triggers without per-event allocation.
	byKind [18][]*configuration.Node
	// resetAt stores timer origin.
	resetAt time.Time
	// timers stores deadline-ordered triggers.
	timers timerQueue
	// unseen stores round-robin effect cursors by stack point.
	unseen map[configuration.Point]int
	// delayed stores outstanding delayed effects for this generation.
	delayed int
	// traces stores a fixed-size trace ring.
	traces [64]Trace
	// traceNext stores the next ring slot.
	traceNext int
	// traceCount stores populated ring entries.
	traceCount int
}

// eventQueue stores one breadth-first stack request.
type eventQueue struct {
	// stack stores the requested stack.
	stack *configuration.Stack
	// trigger stores the matched trigger when the request originated from an event.
	trigger *configuration.Node
	// depth stores call-stack depth.
	depth int
}

// execution stores one trace's bounded mutable work state.
type execution struct {
	// context stores cancellation and deadlines.
	context context.Context
	// state stores the active room state.
	state *state
	// event stores immutable trigger context.
	event trigger.Event
	// now stores injected execution time.
	now time.Time
	// queue stores breadth-first stack requests.
	queue []eventQueue
	// visited stores stacks already executed in this trace.
	visited map[configuration.Point]struct{}
	// trace stores execution counters.
	trace Trace
}

// Metrics returns a consistent-enough lock-free runtime counter snapshot.
func (engine *Engine) Metrics() MetricsSnapshot {
	var snapshot MetricsSnapshot
	for index := range snapshot.Events {
		snapshot.Events[index] = engine.metrics.events[index].Load()
	}
	for index := range snapshot.StackResults {
		snapshot.StackResults[index] = engine.metrics.stackResults[index].Load()
	}
	for index := range snapshot.EffectResults {
		snapshot.EffectResults[index] = engine.metrics.effectResults[index].Load()
	}
	snapshot.BudgetExhausted = engine.metrics.budgetExhausted.Load()
	snapshot.CompileFailures = engine.metrics.compileFailures.Load()
	snapshot.DelayedTasks = engine.metrics.delayedTasks.Load()
	snapshot.CompileCount = engine.metrics.compileCount.Load()
	snapshot.CompileNanoseconds = engine.metrics.compileNanoseconds.Load()
	snapshot.TraceCount = engine.metrics.traceCount.Load()
	snapshot.TraceNanoseconds = engine.metrics.traceNanoseconds.Load()
	return snapshot
}

// recordEffect increments one bounded status counter.
func (engine *Engine) recordEffect(status effect.Status) {
	if int(status) < len(engine.metrics.effectResults) {
		engine.metrics.effectResults[status].Add(1)
	}
}

// recordTrace increments trace duration and budget counters.
func (engine *Engine) recordTrace(trace Trace) {
	engine.metrics.traceCount.Add(1)
	engine.metrics.traceNanoseconds.Add(uint64(trace.Duration))
	if trace.BudgetExhausted {
		engine.metrics.budgetExhausted.Add(1)
	}
}
