// Package tests verifies bounded deterministic WIRED runtime execution.
package tests

import (
	"context"
	"testing"
	"time"

	roomwired "github.com/niflaot/pixels/internal/realm/room/world/wired"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/effect"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
	wiredruntime "github.com/niflaot/pixels/internal/realm/room/world/wired/runtime"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// store returns an immutable test generation source.
type store struct {
	// records stores configured nodes.
	records []record.Config
}

// LoadRoom returns configured nodes.
func (value store) LoadRoom(context.Context, int64) ([]record.Config, error) {
	return append([]record.Config(nil), value.records...), nil
}

// Find is unused by runtime tests.
func (store) Find(context.Context, int64, int64) (record.Config, bool, error) {
	return record.Config{}, false, nil
}

// Save is unused by runtime tests.
func (store) Save(context.Context, record.Config, int64) (record.Config, error) {
	return record.Config{}, nil
}

// SaveRewardConfig is unused by runtime tests.
func (store) SaveRewardConfig(context.Context, record.Config, int64, []record.Reward) (record.Config, error) {
	return record.Config{}, nil
}

// CleanupItem is unused by runtime tests.
func (store) CleanupItem(context.Context, int64) error { return nil }

// Capture is unused by runtime tests.
func (store) Capture(context.Context, int64, int64) ([]record.Target, error) { return nil, nil }

// avatar records player-facing effects.
type avatar struct {
	// calls stores effect count.
	calls int
	// items stores stable executed effect ids.
	items []int64
}

// ExecuteAvatar records one effect.
func (service *avatar) ExecuteAvatar(_ context.Context, _ effect.AvatarOperation, node *configuration.Node, _ trigger.Event) (effect.Result, error) {
	service.calls++
	service.items = append(service.items, node.ItemID)
	return effect.Result{Status: effect.Applied}, nil
}

// scheduler captures one delayed callback for deterministic execution.
type scheduler struct {
	// delay stores the scheduled duration.
	delay time.Duration
	// run stores the scheduled callback.
	run func(time.Time)
}

// Schedule captures delayed work.
func (value *scheduler) Schedule(_ int64, _ uint64, delay time.Duration, run func(time.Time)) bool {
	value.delay, value.run = delay, run
	return true
}

// TestCallStacksUsesBreadthFirstVisitedTrace verifies call stacks and loop prevention.
func TestCallStacksUsesBreadthFirstVisitedTrace(t *testing.T) {
	records := []record.Config{
		{ItemID: 1, RoomID: 1, Interaction: "wf_trg_enter_room", X: 1, Y: 1},
		{ItemID: 2, RoomID: 1, Interaction: "wf_act_call_stacks", X: 1, Y: 1, SelectionMode: 1, Targets: []record.Target{{ItemID: 3}}},
		{ItemID: 3, RoomID: 1, Interaction: "wf_act_show_message", X: 2, Y: 2, StringParam: "hello"},
		{ItemID: 4, RoomID: 1, Interaction: "wf_act_call_stacks", X: 2, Y: 2, SelectionMode: 1, Targets: []record.Target{{ItemID: 1}}},
	}
	service := &avatar{}
	engine := newEngine(t, records, service)
	now := time.Now()
	if err := engine.Reload(context.Background(), 1, now); err != nil {
		t.Fatal(err)
	}
	trace, err := engine.Process(context.Background(), trigger.Event{Kind: trigger.EnterRoom, RoomID: 1, ActorKind: trigger.ActorPlayer, PlayerID: 9}, now)
	if err != nil {
		t.Fatalf("Process() error = %v", err)
	}
	if trace.Stacks != 2 || trace.Effects != 3 || service.calls != 1 || trace.BudgetExhausted {
		t.Fatalf("trace=%+v calls=%d", trace, service.calls)
	}
}

// TestCycleFiresAtMostOneMissedPeriod verifies timer catch-up policy.
func TestCycleFiresAtMostOneMissedPeriod(t *testing.T) {
	records := []record.Config{
		{ItemID: 1, RoomID: 1, Interaction: "wf_trg_periodically", X: 1, Y: 1, IntParams: []int32{1}},
		{ItemID: 2, RoomID: 1, Interaction: "wf_act_show_message", X: 1, Y: 1, StringParam: "tick"},
	}
	service := &avatar{}
	engine := newEngine(t, records, service)
	now := time.Now()
	if err := engine.Reload(context.Background(), 1, now); err != nil {
		t.Fatal(err)
	}
	if err := engine.Cycle(context.Background(), 1, now.Add(10*time.Second)); err != nil {
		t.Fatal(err)
	}
	if service.calls != 1 {
		t.Fatalf("timer calls = %d", service.calls)
	}
	if err := engine.Cycle(context.Background(), 1, now.Add(10*time.Second)); err != nil {
		t.Fatal(err)
	}
	if service.calls != 1 {
		t.Fatalf("same deadline calls = %d", service.calls)
	}
}

// TestNoCandidatesDoesNoWork verifies unrelated room events stay empty.
func TestNoCandidatesDoesNoWork(t *testing.T) {
	engine := newEngine(t, []record.Config{{ItemID: 1, RoomID: 1, Interaction: "wf_trg_enter_room", X: 1, Y: 1}}, &avatar{})
	now := time.Now()
	if err := engine.Reload(context.Background(), 1, now); err != nil {
		t.Fatal(err)
	}
	trace, err := engine.Process(context.Background(), trigger.Event{Kind: trigger.StateChanged, RoomID: 1}, now)
	if err != nil || trace != (wiredruntime.Trace{}) {
		t.Fatalf("trace=%+v err=%v", trace, err)
	}
}

// TestDelayedEffectsRevalidateGeneration verifies callbacks cannot escape reload or close.
func TestDelayedEffectsRevalidateGeneration(t *testing.T) {
	records := []record.Config{
		{ItemID: 1, RoomID: 1, Interaction: "wf_trg_enter_room", X: 1, Y: 1},
		{ItemID: 2, RoomID: 1, Interaction: "wf_act_show_message", X: 1, Y: 1, StringParam: "later", DelayPulses: 2},
	}
	service, delayed := &avatar{}, &scheduler{}
	engine := newEngineWithScheduler(t, records, service, delayed)
	now := time.Now()
	if err := engine.Reload(context.Background(), 1, now); err != nil {
		t.Fatal(err)
	}
	if _, err := engine.Process(context.Background(), trigger.Event{Kind: trigger.EnterRoom, RoomID: 1, ActorKind: trigger.ActorPlayer, PlayerID: 9}, now); err != nil {
		t.Fatal(err)
	}
	if service.calls != 0 || delayed.run == nil || delayed.delay != time.Second {
		t.Fatalf("calls=%d delay=%v callback=%v", service.calls, delayed.delay, delayed.run != nil)
	}
	delayed.run(now.Add(time.Second))
	if service.calls != 1 {
		t.Fatalf("active delayed calls=%d", service.calls)
	}
	_, _ = engine.Process(context.Background(), trigger.Event{Kind: trigger.EnterRoom, RoomID: 1, ActorKind: trigger.ActorPlayer, PlayerID: 9}, now)
	stale := delayed.run
	if err := engine.Reload(context.Background(), 1, now); err != nil {
		t.Fatal(err)
	}
	stale(now.Add(time.Second))
	if service.calls != 1 {
		t.Fatalf("stale generation executed: calls=%d", service.calls)
	}
	_, _ = engine.Process(context.Background(), trigger.Event{Kind: trigger.EnterRoom, RoomID: 1, ActorKind: trigger.ActorPlayer, PlayerID: 9}, now)
	closed := delayed.run
	engine.Close(1)
	closed(now.Add(time.Second))
	if service.calls != 1 {
		t.Fatalf("closed generation executed: calls=%d", service.calls)
	}
}

// TestUnseenRotatesAndReloadResets verifies deterministic round-robin runtime state.
func TestUnseenRotatesAndReloadResets(t *testing.T) {
	records := []record.Config{
		{ItemID: 1, RoomID: 1, Interaction: "wf_trg_enter_room", X: 1, Y: 1},
		{ItemID: 2, RoomID: 1, Interaction: "wf_act_show_message", X: 1, Y: 1, StringParam: "a"},
		{ItemID: 3, RoomID: 1, Interaction: "wf_act_show_message", X: 1, Y: 1, StringParam: "b"},
		{ItemID: 4, RoomID: 1, Interaction: "wf_xtra_unseen", X: 1, Y: 1},
	}
	service := &avatar{}
	engine := newEngine(t, records, service)
	now := time.Now()
	_ = engine.Reload(context.Background(), 1, now)
	for index := 0; index < 3; index++ {
		_, _ = engine.Process(context.Background(), trigger.Event{Kind: trigger.EnterRoom, RoomID: 1, ActorKind: trigger.ActorPlayer, PlayerID: 9}, now)
	}
	want := []int64{2, 3, 2}
	for index := range want {
		if service.items[index] != want[index] {
			t.Fatalf("unseen items=%v", service.items)
		}
	}
	_ = engine.Reload(context.Background(), 1, now)
	_, _ = engine.Process(context.Background(), trigger.Event{Kind: trigger.EnterRoom, RoomID: 1, ActorKind: trigger.ActorPlayer, PlayerID: 9}, now)
	if service.items[len(service.items)-1] != 2 {
		t.Fatalf("reload did not reset unseen: %v", service.items)
	}
}

// newEngine creates a canonical runtime test engine.
func newEngine(t testing.TB, records []record.Config, service *avatar) *wiredruntime.Engine {
	return newEngineWithScheduler(t, records, service, nil)
}

// newEngineWithScheduler creates a runtime test engine with delayed work capture.
func newEngineWithScheduler(t testing.TB, records []record.Config, service *avatar, delayed wiredruntime.Scheduler) *wiredruntime.Engine {
	t.Helper()
	registered, err := registry.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	compiler := configuration.NewCompiler(registered, roomwired.Config{})
	executor := effect.New(effect.Services{Avatar: service})
	return wiredruntime.New(roomwired.Config{Enabled: true}, store{records: records}, compiler, executor, nil, delayed, nil)
}
