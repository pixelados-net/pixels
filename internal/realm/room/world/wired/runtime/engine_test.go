package runtime

import (
	"context"
	"errors"
	"testing"
	"time"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	roomwired "github.com/niflaot/pixels/internal/realm/room/world/wired"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/effect"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
)

// engineStore returns immutable runtime test records.
type engineStore struct {
	// records stores one room generation.
	records []record.Config
}

// LoadRoom returns immutable configuration copies.
func (store engineStore) LoadRoom(context.Context, int64) ([]record.Config, error) {
	return append([]record.Config(nil), store.records...), nil
}

// Find is unused by runtime package tests.
func (engineStore) Find(context.Context, int64, int64) (record.Config, bool, error) {
	return record.Config{}, false, nil
}

// Save is unused by runtime package tests.
func (engineStore) Save(context.Context, record.Config, int64) (record.Config, error) {
	return record.Config{}, nil
}

// Capture is unused by runtime package tests.
func (engineStore) Capture(context.Context, int64, int64) ([]record.Target, error) { return nil, nil }

// SaveRewardConfig is unused by runtime package tests.
func (engineStore) SaveRewardConfig(context.Context, record.Config, int64, []record.Reward) (record.Config, error) {
	return record.Config{}, nil
}

// CleanupItem is unused by runtime package tests.
func (engineStore) CleanupItem(context.Context, int64) error { return nil }

// TestEngineInspectionAndTimerControls verifies runtime diagnostics and lifecycle controls.
func TestEngineInspectionAndTimerControls(t *testing.T) {
	records := []record.Config{
		{ItemID: 1, RoomID: 1, Interaction: "wf_trg_enter_room", SpriteID: 10, X: 1, Y: 1},
		{ItemID: 2, RoomID: 1, Interaction: "wf_act_show_message", SpriteID: 20, X: 1, Y: 1, StringParam: "hello"},
		{ItemID: 3, RoomID: 1, Interaction: "wf_trg_periodically", SpriteID: 30, X: 2, Y: 2, IntParams: []int32{1}},
		{ItemID: 4, RoomID: 1, Interaction: "wf_act_show_message", SpriteID: 40, X: 2, Y: 2, StringParam: "tick"},
	}
	engine := runtimeEngine(t, records)
	now := time.Now()
	if err := engine.Reload(context.Background(), 1, now); err != nil {
		t.Fatal(err)
	}
	if !engine.Loaded(1) || !engine.IsCurrent(1, 1) {
		t.Fatal("reloaded generation was not current")
	}
	event := trigger.Event{Kind: trigger.EnterRoom, RoomID: 1, ActorKind: trigger.ActorPlayer, ActorID: 7, PlayerID: 7}
	if !engine.Matches(event) || engine.Matches(trigger.Event{Kind: trigger.Collision, RoomID: 1}) {
		t.Fatal("trigger candidate matching failed")
	}
	trace, err := engine.Process(context.Background(), event, now)
	if err != nil || trace.Stacks != 1 {
		t.Fatalf("trace=%+v err=%v", trace, err)
	}
	if traces := engine.Traces(1); len(traces) != 1 || traces[0].ID != trace.ID {
		t.Fatalf("trace ring=%+v", traces)
	}
	if !engine.ResetTimers(1, now.Add(time.Second)) || engine.ResetTimers(999, now) {
		t.Fatal("timer reset lifecycle failed")
	}
	if conflicts := engine.Conflicts(1, 4); len(conflicts) != 1 || conflicts[0] != 30 {
		t.Fatalf("actor conflicts=%v", conflicts)
	}
	metrics := engine.Metrics()
	if metrics.TraceCount == 0 || metrics.CompileCount == 0 {
		t.Fatalf("metrics=%+v", metrics)
	}
	engine.Close(1)
	if engine.Loaded(1) || engine.Traces(1) != nil {
		t.Fatal("closed engine retained room state")
	}
}

// TestRuntimeHelpersCoverStableOrdering verifies timer, event, error, and random helper policies.
func TestRuntimeHelpersCoverStableOrdering(t *testing.T) {
	now := time.Now()
	left := &configuration.Node{ItemID: 1}
	right := &configuration.Node{ItemID: 2}
	queue := timerQueue{{node: right, deadline: now}, {node: left, deadline: now}}
	if !queue.Less(1, 0) {
		t.Fatal("equal timer deadlines were not ordered by item id")
	}
	queue.Swap(0, 1)
	queue.Push(timerEntry{node: right, deadline: now.Add(time.Second)})
	if queue.Len() != 3 || queue.Pop().(timerEntry).node != right {
		t.Fatal("timer heap primitives failed")
	}
	keys := []string{"wf_trg_enter_room", "wf_trg_says_something", "wf_trg_walks_on_furni", "wf_trg_walks_off_furni", "wf_trg_state_changed", "wf_trg_collision", "wf_trg_game_starts", "wf_trg_game_ends", "wf_trg_score_achieved", "wf_trg_bot_reached_stf", "wf_trg_bot_reached_avtr", "wf_trg_game_team_win", "wf_trg_game_team_lose", "wf_trg_periodically", "wf_trg_period_long", "wf_trg_at_given_time", "wf_trg_at_time_long"}
	for _, key := range keys {
		if eventKind(key) == 0 {
			t.Errorf("event kind missing for %s", key)
		}
	}
	if eventKind("unknown") != 0 || randomIndex(1, configuration.Point{}, 1) != 0 {
		t.Fatal("helper defaults failed")
	}
	if first := randomIndex(7, configuration.Point{X: 2, Y: 3}, 10); first != randomIndex(7, configuration.Point{X: 2, Y: 3}, 10) {
		t.Fatal("random selector was not deterministic")
	}
	leftErr, rightErr := errors.New("left"), errors.New("right")
	if err := joined(leftErr, rightErr); !errors.Is(err, leftErr) || !errors.Is(err, rightErr) {
		t.Fatalf("joined error=%v", err)
	}
}

// TestRoomSchedulerQueuesOnlyActiveRooms verifies delayed work stays owned by active room lifecycle.
func TestRoomSchedulerQueuesOnlyActiveRooms(t *testing.T) {
	rooms := roomlive.NewRegistry(nil)
	scheduler := NewRoomScheduler(rooms)
	if scheduler.Schedule(1, 1, 0, nil) || scheduler.Schedule(1, 1, 0, func(time.Time) {}) {
		t.Fatal("invalid scheduler input was accepted")
	}
	active, err := rooms.Activate(roomlive.Snapshot{ID: 1, MaxUsers: 2})
	if err != nil {
		t.Fatal(err)
	}
	called := false
	if !scheduler.Schedule(1, 1, 0, func(time.Time) { called = true }) {
		t.Fatal("active room schedule was rejected")
	}
	active.RunScheduled(time.Now().Add(time.Second))
	if !called {
		t.Fatal("scheduled callback did not run")
	}
	_, _, _ = rooms.Close(context.Background(), 1)
}

// runtimeEngine creates an enabled engine from immutable test records.
func runtimeEngine(t *testing.T, records []record.Config) *Engine {
	t.Helper()
	registered, err := registry.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	compiler := configuration.NewCompiler(registered, roomwired.Config{})
	return New(roomwired.Config{Enabled: true}, engineStore{records: records}, compiler, effect.New(effect.Services{}), nil, nil, nil)
}
