package wiring

import (
	"context"
	"testing"
	"time"

	socialgroup "github.com/niflaot/pixels/internal/realm/group"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomentered "github.com/niflaot/pixels/internal/realm/room/access/events/entered"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/game"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/fx"
)

// wiringGroups returns one deterministic room membership snapshot.
type wiringGroups struct{}

// RoomMembership returns one linked test member.
func (wiringGroups) RoomMembership(context.Context, int64) (int64, []int64, bool, error) {
	return 1, []int64{7}, true, nil
}

// wiringLifecycle captures the registered shutdown hook.
type wiringLifecycle struct {
	// hooks stores appended lifecycle hooks.
	hooks []fx.Hook
}

// Append captures one lifecycle hook.
func (lifecycle *wiringLifecycle) Append(hook fx.Hook) {
	lifecycle.hooks = append(lifecycle.hooks, hook)
}

// TestEnteredHandlerWarmsDependenciesAndSchedulesTrigger verifies cold room entry bootstrap.
func TestEnteredHandlerWarmsDependenciesAndSchedulesTrigger(t *testing.T) {
	rooms, active := wiringRoom(t)
	store := &wiringStore{records: []record.Config{
		{ItemID: 1, RoomID: active.ID(), Interaction: "wf_trg_enter_room", X: 1, Y: 1},
		{ItemID: 2, RoomID: active.ID(), Interaction: "wf_act_show_message", X: 1, Y: 1, StringParam: "entered"},
	}}
	avatar := &wiringAvatar{}
	engine := wiringEngine(t, store, avatar)
	groups := socialgroup.New(wiringGroups{})
	handler := enteredHandler(rooms, playerlive.NewRegistry(), engine, groups, nil)
	if err := handler(context.Background(), bus.Event{Payload: "foreign"}); err != nil {
		t.Fatal(err)
	}
	if err := handler(context.Background(), bus.Event{Payload: roomentered.Payload{RoomID: active.ID(), PlayerID: 7}}); err != nil {
		t.Fatal(err)
	}
	active.RunScheduled(time.Now().Add(time.Second))
	if avatar.calls != 1 || !engine.Loaded(active.ID()) {
		t.Fatalf("calls=%d loaded=%t", avatar.calls, engine.Loaded(active.ID()))
	}
	if member, loaded := groups.IsRoomMember(active.ID(), 7); !loaded || !member {
		t.Fatal("room social membership was not warmed")
	}
}

// TestRegisterRuntimeSubscribesAndReleasesLifecycle verifies subscriptions and close cleanup.
func TestRegisterRuntimeSubscribesAndReleasesLifecycle(t *testing.T) {
	rooms, active := wiringRoom(t)
	store := &wiringStore{}
	engine := wiringEngine(t, store, &wiringAvatar{})
	if err := engine.Reload(context.Background(), active.ID(), time.Now()); err != nil {
		t.Fatal(err)
	}
	groups := socialgroup.New(wiringGroups{})
	if err := groups.PrepareRoom(context.Background(), active.ID()); err != nil {
		t.Fatal(err)
	}
	lifecycle := &wiringLifecycle{}
	local := bus.New()
	if err := RegisterRuntime(lifecycle, local, rooms, playerlive.NewRegistry(), nil, engine, game.New(), nil, groups, store, nil); err != nil {
		t.Fatal(err)
	}
	if len(lifecycle.hooks) != 1 || lifecycle.hooks[0].OnStop == nil {
		t.Fatalf("lifecycle hooks=%d", len(lifecycle.hooks))
	}
	if err := lifecycle.hooks[0].OnStop(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, _, err := rooms.Close(context.Background(), active.ID()); err != nil {
		t.Fatal(err)
	}
	if engine.Loaded(active.ID()) {
		t.Fatal("room close retained WIRED generation")
	}
	if _, loaded := groups.IsRoomMember(active.ID(), 7); loaded {
		t.Fatal("room close retained group snapshot")
	}
}
