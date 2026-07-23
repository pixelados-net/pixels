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

// activation records executed box projections.
type activation struct {
	// items stores projected stable item identifiers.
	items []int64
}

// Activate records one activation projection.
func (value *activation) Activate(_ context.Context, _ int64, itemID int64) error {
	value.items = append(value.items, itemID)
	return nil
}

// TestActivationProjectsOnlyExecutedBoxes verifies canonical trigger and effect animation.
func TestActivationProjectsOnlyExecutedBoxes(t *testing.T) {
	records := []record.Config{
		{ItemID: 2, RoomID: 1, Interaction: "wf_act_show_message", X: 1, Y: 1, StringParam: "hello"},
		{ItemID: 1, RoomID: 1, Interaction: "wf_trg_enter_room", X: 1, Y: 1},
	}
	projector, delayed := &activation{}, &scheduler{}
	engine := configuredEngine(t, roomwired.Config{Enabled: true}, records, delayed, projector)
	now := time.Now()
	if err := engine.Reload(context.Background(), 1, now); err != nil {
		t.Fatal(err)
	}
	_, _ = engine.Process(context.Background(), trigger.Event{Kind: trigger.EnterRoom, RoomID: 1, ActorKind: trigger.ActorPlayer, PlayerID: 9}, now)
	if len(projector.items) != 2 || projector.items[0] != 1 || projector.items[1] != 2 {
		t.Fatalf("activation items=%v", projector.items)
	}
}

// TestFailedConditionActivatesOnlyTrigger verifies evaluated conditions never animate as effects.
func TestFailedConditionActivatesOnlyTrigger(t *testing.T) {
	records := []record.Config{
		{ItemID: 1, RoomID: 1, Interaction: "wf_trg_enter_room", X: 1, Y: 1},
		{ItemID: 2, RoomID: 1, Interaction: "wf_cnd_date_rng_active", X: 1, Y: 1, IntParams: []int32{0, 2147483647}},
		{ItemID: 3, RoomID: 1, Interaction: "wf_act_show_message", X: 1, Y: 1, StringParam: "blocked"},
	}
	projector := &activation{}
	engine := configuredEngine(t, roomwired.Config{Enabled: true}, records, &scheduler{}, projector)
	now := time.Now()
	if err := engine.Reload(context.Background(), 1, now); err != nil {
		t.Fatal(err)
	}
	_, _ = engine.Process(context.Background(), trigger.Event{Kind: trigger.EnterRoom, RoomID: 1, ActorKind: trigger.ActorPlayer, PlayerID: 9}, now)
	if len(projector.items) != 1 || projector.items[0] != 1 {
		t.Fatalf("failed-condition activation items=%v", projector.items)
	}
}

// TestSelectorActivatesOnlyChosenEffect verifies extras and one selected branch animate.
func TestSelectorActivatesOnlyChosenEffect(t *testing.T) {
	records := []record.Config{
		{ItemID: 1, RoomID: 1, Interaction: "wf_trg_enter_room", X: 1, Y: 1},
		{ItemID: 2, RoomID: 1, Interaction: "wf_xtra_random", X: 1, Y: 1},
		{ItemID: 3, RoomID: 1, Interaction: "wf_act_show_message", X: 1, Y: 1, StringParam: "a"},
		{ItemID: 4, RoomID: 1, Interaction: "wf_act_show_message", X: 1, Y: 1, StringParam: "b"},
	}
	projector := &activation{}
	engine := configuredEngine(t, roomwired.Config{Enabled: true}, records, &scheduler{}, projector)
	now := time.Now()
	if err := engine.Reload(context.Background(), 1, now); err != nil {
		t.Fatal(err)
	}
	_, _ = engine.Process(context.Background(), trigger.Event{ID: 7, Kind: trigger.EnterRoom, RoomID: 1, ActorKind: trigger.ActorPlayer, PlayerID: 9}, now)
	if len(projector.items) != 3 || projector.items[0] != 1 || projector.items[1] != 2 {
		t.Fatalf("selector activation items=%v", projector.items)
	}
	if projector.items[2] != 3 && projector.items[2] != 4 {
		t.Fatalf("unexpected selected effect=%d", projector.items[2])
	}
}

// TestDelayedEffectActivatesOnExecution verifies delay scheduling does not animate early.
func TestDelayedEffectActivatesOnExecution(t *testing.T) {
	records := []record.Config{
		{ItemID: 1, RoomID: 1, Interaction: "wf_trg_enter_room", X: 1, Y: 1},
		{ItemID: 2, RoomID: 1, Interaction: "wf_act_show_message", X: 1, Y: 1, StringParam: "later", DelayPulses: 1},
	}
	projector, delayed := &activation{}, &scheduler{}
	engine := configuredEngine(t, roomwired.Config{Enabled: true}, records, delayed, projector)
	now := time.Now()
	if err := engine.Reload(context.Background(), 1, now); err != nil {
		t.Fatal(err)
	}
	_, _ = engine.Process(context.Background(), trigger.Event{Kind: trigger.EnterRoom, RoomID: 1, ActorKind: trigger.ActorPlayer, PlayerID: 9}, now)
	if len(projector.items) != 1 || projector.items[0] != 1 {
		t.Fatalf("early activation items=%v", projector.items)
	}
	delayed.run(now.Add(500 * time.Millisecond))
	if len(projector.items) != 2 || projector.items[1] != 2 {
		t.Fatalf("delayed activation items=%v", projector.items)
	}
}

// TestDelayedBudgetAndMetricsResetOnClose verifies bounded tasks and lifecycle gauge cleanup.
func TestDelayedBudgetAndMetricsResetOnClose(t *testing.T) {
	records := []record.Config{
		{ItemID: 1, RoomID: 1, Interaction: "wf_trg_enter_room", X: 1, Y: 1},
		{ItemID: 2, RoomID: 1, Interaction: "wf_act_show_message", X: 1, Y: 1, StringParam: "a", DelayPulses: 1},
		{ItemID: 3, RoomID: 1, Interaction: "wf_act_show_message", X: 1, Y: 1, StringParam: "b", DelayPulses: 1},
	}
	delayed := &scheduler{}
	engine := configuredEngine(t, roomwired.Config{Enabled: true, MaxDelayedPerRoom: 1}, records, delayed, nil)
	now := time.Now()
	_ = engine.Reload(context.Background(), 1, now)
	_, _ = engine.Process(context.Background(), trigger.Event{Kind: trigger.EnterRoom, RoomID: 1, ActorKind: trigger.ActorPlayer, PlayerID: 9}, now)
	metrics := engine.Metrics()
	if metrics.DelayedTasks != 1 || metrics.Events[trigger.EnterRoom] != 1 || metrics.TraceCount != 1 {
		t.Fatalf("metrics=%+v", metrics)
	}
	engine.Close(1)
	if metrics = engine.Metrics(); metrics.DelayedTasks != 0 {
		t.Fatalf("close metrics=%+v", metrics)
	}
}

// configuredEngine creates an engine with explicit runtime collaborators.
func configuredEngine(t testing.TB, config roomwired.Config, records []record.Config, delayed wiredruntime.Scheduler, projector wiredruntime.Activator) *wiredruntime.Engine {
	t.Helper()
	registered, err := registry.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	compiler := configuration.NewCompiler(registered, config)
	executor := effect.New(effect.Services{Avatar: &avatar{}})
	return wiredruntime.New(config, store{records: records}, compiler, executor, nil, delayed, projector)
}
