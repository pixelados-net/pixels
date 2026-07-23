package wiring

import (
	"context"
	"testing"
	"time"

	furnituremoved "github.com/niflaot/pixels/internal/realm/furniture/events/moved"
	furniturepicked "github.com/niflaot/pixels/internal/realm/furniture/events/pickedup"
	furnitureplaced "github.com/niflaot/pixels/internal/realm/furniture/events/placed"
	furnitureused "github.com/niflaot/pixels/internal/realm/furniture/events/used"
	furnitureoff "github.com/niflaot/pixels/internal/realm/furniture/events/walkedoff"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	worldfurniture "github.com/niflaot/pixels/internal/realm/room/world/furniture"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	roomwired "github.com/niflaot/pixels/internal/realm/room/world/wired"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/effect"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
	wiredruntime "github.com/niflaot/pixels/internal/realm/room/world/wired/runtime"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
	"github.com/niflaot/pixels/pkg/bus"
)

// wiringStore stores immutable handler fixtures and cleanup observations.
type wiringStore struct {
	// records stores one compiled generation.
	records []record.Config
	// cleaned stores the last picked-up item.
	cleaned int64
}

// LoadRoom returns immutable handler fixtures.
func (store *wiringStore) LoadRoom(context.Context, int64) ([]record.Config, error) {
	return append([]record.Config(nil), store.records...), nil
}

// Find is unused by wiring tests.
func (*wiringStore) Find(context.Context, int64, int64) (record.Config, bool, error) {
	return record.Config{}, false, nil
}

// Save is unused by wiring tests.
func (*wiringStore) Save(context.Context, record.Config, int64) (record.Config, error) {
	return record.Config{}, nil
}

// Capture is unused by wiring tests.
func (*wiringStore) Capture(context.Context, int64, int64) ([]record.Target, error) { return nil, nil }

// SaveRewardConfig is unused by wiring tests.
func (*wiringStore) SaveRewardConfig(context.Context, record.Config, int64, []record.Reward) (record.Config, error) {
	return record.Config{}, nil
}

// CleanupItem observes picked-up WIRED cleanup.
func (store *wiringStore) CleanupItem(_ context.Context, itemID int64) error {
	store.cleaned = itemID
	return nil
}

// wiringAvatar counts executed handler stacks.
type wiringAvatar struct {
	// calls stores executed effects.
	calls int
}

// ExecuteAvatar observes one executed stack.
func (avatar *wiringAvatar) ExecuteAvatar(context.Context, effect.AvatarOperation, *configuration.Node, trigger.Event) (effect.Result, error) {
	avatar.calls++
	return effect.Result{Status: effect.Applied}, nil
}

// TestEventHandlersScheduleAndReloadActiveRoom verifies hot events serialize through room tasks.
func TestEventHandlersScheduleAndReloadActiveRoom(t *testing.T) {
	rooms, active := wiringRoom(t)
	records := []record.Config{
		{ItemID: 1, RoomID: active.ID(), Interaction: "wf_trg_state_changed", X: 1, Y: 1, SelectionMode: 1, Targets: []record.Target{{ItemID: 10}}},
		{ItemID: 2, RoomID: active.ID(), Interaction: "wf_act_show_message", X: 1, Y: 1, StringParam: "state"},
		{ItemID: 3, RoomID: active.ID(), Interaction: "wf_trg_walks_off_furni", X: 2, Y: 2, SelectionMode: 1, Targets: []record.Target{{ItemID: 10}}},
		{ItemID: 4, RoomID: active.ID(), Interaction: "wf_act_show_message", X: 2, Y: 2, StringParam: "off"},
		{ItemID: 5, RoomID: active.ID(), Interaction: "wf_trg_collision", X: 3, Y: 3},
		{ItemID: 6, RoomID: active.ID(), Interaction: "wf_act_show_message", X: 3, Y: 3, StringParam: "collision"},
	}
	store := &wiringStore{records: records}
	avatar := &wiringAvatar{}
	engine := wiringEngine(t, store, avatar)
	if err := engine.Reload(context.Background(), active.ID(), time.Now()); err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := usedHandler(rooms, engine)(ctx, bus.Event{Payload: furnitureused.Payload{RoomID: active.ID(), PlayerID: 7, ItemID: 10}}); err != nil {
		t.Fatal(err)
	}
	active.RunScheduled(time.Now().Add(time.Second))
	if err := walkOffHandler(rooms, engine)(ctx, bus.Event{Payload: furnitureoff.Payload{RoomID: active.ID(), PlayerID: 7, ItemID: 10}}); err != nil {
		t.Fatal(err)
	}
	active.RunScheduled(time.Now().Add(time.Second))
	if avatar.calls != 2 {
		t.Fatalf("handler effect calls=%d, want 2", avatar.calls)
	}
	if event := unitEvent(rooms, trigger.StateChanged, active.ID(), 7, 10); event.SourceSprite != 2 || event.ActorKind != trigger.ActorPlayer {
		t.Fatalf("unit event=%+v", event)
	}
	before := engine.Metrics().CompileCount
	_ = reloadMovedHandler(engine, nil)(ctx, bus.Event{Payload: furnituremoved.Payload{RoomID: active.ID(), ItemID: 10}})
	_ = reloadPlacedHandler(engine, nil)(ctx, bus.Event{Payload: furnitureplaced.Payload{RoomID: active.ID(), ItemID: 10}})
	_ = reloadPickedHandler(store, engine, nil)(ctx, bus.Event{Payload: furniturepicked.Payload{RoomID: active.ID(), ItemID: 10}})
	if engine.Metrics().CompileCount != before+3 || store.cleaned != 10 {
		t.Fatalf("compile count=%d cleanup=%d", engine.Metrics().CompileCount, store.cleaned)
	}
}

// TestEventHandlersIgnoreForeignPayloadsAndWiredBoxUse verifies adapters fail closed.
func TestEventHandlersIgnoreForeignPayloadsAndWiredBoxUse(t *testing.T) {
	rooms, active := wiringRoom(t)
	store, avatar := &wiringStore{}, &wiringAvatar{}
	engine := wiringEngine(t, store, avatar)
	if err := engine.Reload(context.Background(), active.ID(), time.Now()); err != nil {
		t.Fatal(err)
	}
	handlers := []bus.Handler{
		usedHandler(rooms, engine), walkOffHandler(rooms, engine),
		reloadMovedHandler(engine, nil), reloadPlacedHandler(engine, nil), reloadPickedHandler(store, engine, nil),
		walkOnHandler(rooms, engine, nil), unitMovedHandler(rooms, nil, engine),
	}
	for _, handler := range handlers {
		if err := handler(context.Background(), bus.Event{Payload: "foreign"}); err != nil {
			t.Fatal(err)
		}
	}
	_ = usedHandler(rooms, engine)(context.Background(), bus.Event{Payload: furnitureused.Payload{RoomID: active.ID(), PlayerID: 7, ItemID: 11}})
	active.RunScheduled(time.Now().Add(time.Second))
	if avatar.calls != 0 {
		t.Fatal("WIRED box activation fed state trigger recursion")
	}
	scheduleEvent(rooms, engine, trigger.Event{RoomID: 999, Kind: trigger.EnterRoom})
	logError(nil, "ignored", active.ID(), nil)
}

// wiringEngine creates one enabled handler test engine.
func wiringEngine(t *testing.T, store record.Store, avatar *wiringAvatar) *wiredruntime.Engine {
	t.Helper()
	registered, err := registry.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	return wiredruntime.New(roomwired.Config{Enabled: true}, store, configuration.NewCompiler(registered, roomwired.Config{}), effect.New(effect.Services{Avatar: avatar}), nil, nil, nil)
}

// wiringRoom creates one player room with regular and WIRED furniture.
func wiringRoom(t *testing.T) (*roomlive.Registry, *roomlive.Room) {
	t.Helper()
	rooms := roomlive.NewRegistry(nil)
	active, err := rooms.Activate(roomlive.Snapshot{ID: 42, MaxUsers: 5})
	if err != nil {
		t.Fatal(err)
	}
	roomGrid, err := grid.Parse("0", grid.WithDoor(0, 0))
	if err != nil {
		t.Fatal(err)
	}
	items := []worldfurniture.Item{
		{ID: 10, Point: grid.MustPoint(0, 0), Definition: worldfurniture.Definition{SpriteID: 2, InteractionType: "toggle", Width: 1, Length: 1, AllowWalk: true}},
		{ID: 11, Point: grid.MustPoint(0, 0), Definition: worldfurniture.Definition{SpriteID: 3, InteractionType: "wf_trg_state_changed", Width: 1, Length: 1, AllowWalk: true}},
	}
	if err = active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Furniture: items, Door: worldpath.Position{Point: grid.MustPoint(0, 0)}}); err != nil {
		t.Fatal(err)
	}
	if _, err = rooms.Join(context.Background(), active.ID(), roomlive.Occupant{PlayerID: 7, ConnectionID: "test", ConnectionKind: "test"}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _, _ = rooms.Close(context.Background(), active.ID()) })
	return rooms, active
}
