package tests

import (
	"context"
	"testing"
	"time"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	wiredruntime "github.com/niflaot/pixels/internal/realm/room/world/wired/runtime"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// BenchmarkNoCandidateDispatch measures the mandatory empty-event fast path.
func BenchmarkNoCandidateDispatch(benchmark *testing.B) {
	engine := newEngineForBenchmark(benchmark, []record.Config{{ItemID: 1, RoomID: 1, Interaction: "wf_trg_enter_room", X: 1, Y: 1}}, &avatar{})
	event := trigger.Event{Kind: trigger.StateChanged, RoomID: 1}
	now := time.Now()
	benchmark.ReportAllocs()
	benchmark.ResetTimer()
	for benchmark.Loop() {
		_, _ = engine.Process(context.Background(), event, now)
	}
}

// BenchmarkIndexedDispatch measures one indexed trigger and effect stack.
func BenchmarkIndexedDispatch(benchmark *testing.B) {
	records := []record.Config{
		{ItemID: 1, RoomID: 1, Interaction: "wf_trg_enter_room", X: 1, Y: 1},
		{ItemID: 2, RoomID: 1, Interaction: "wf_act_show_message", X: 1, Y: 1, StringParam: "hello"},
	}
	engine := newEngineForBenchmark(benchmark, records, &avatar{})
	event := trigger.Event{Kind: trigger.EnterRoom, RoomID: 1, ActorKind: trigger.ActorPlayer, PlayerID: 4}
	now := time.Now()
	benchmark.ReportAllocs()
	benchmark.ResetTimer()
	for benchmark.Loop() {
		_, _ = engine.Process(context.Background(), event, now)
	}
}

// newEngineForBenchmark adapts the test helper to benchmarks.
func newEngineForBenchmark(benchmark *testing.B, records []record.Config, service *avatar) *wiredruntime.Engine {
	benchmark.Helper()
	engine := newEngine(benchmark, records, service)
	if err := engine.Reload(context.Background(), 1, time.Now()); err != nil {
		benchmark.Fatal(err)
	}
	return engine
}
