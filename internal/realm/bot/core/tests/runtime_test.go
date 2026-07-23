// Package tests exercises the public bot core without expanding its six-file package.
package tests

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	botbehavior "github.com/niflaot/pixels/internal/realm/bot/behavior"
	botcore "github.com/niflaot/pixels/internal/realm/bot/core"
	botpolicy "github.com/niflaot/pixels/internal/realm/bot/policy"
	botrecord "github.com/niflaot/pixels/internal/realm/bot/record"
	playerlive "github.com/niflaot/pixels/internal/realm/player/live"
	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	"github.com/niflaot/pixels/internal/realm/room/world/grid"
	worldpath "github.com/niflaot/pixels/internal/realm/room/world/path"
	worldunit "github.com/niflaot/pixels/internal/realm/room/world/unit"
	sdkbot "github.com/niflaot/pixels/sdk/bot"
	"go.uber.org/zap"
)

// roomStore supplies only the persistence operation used by core room loading.
type roomStore struct {
	botrecord.Store
	// bots stores placed test records.
	bots []botrecord.Bot
	// serveItems stores bartender mappings.
	serveItems []botrecord.ServeItem
	// visits stores visitor-log rows.
	visits []botrecord.Visit
}

// Room returns configured placed bots.
func (store *roomStore) Room(context.Context, int64) ([]botrecord.Bot, error) { return store.bots, nil }

// ListServeItems returns configured bartender mappings.
func (store *roomStore) ListServeItems(context.Context) ([]botrecord.ServeItem, error) {
	return store.serveItems, nil
}

// RecordVisit accepts one test room entry.
func (*roomStore) RecordVisit(context.Context, int64, int64) error { return nil }

// VisitsSince returns configured visitor rows.
func (store *roomStore) VisitsSince(context.Context, int64, int64, time.Time, int) ([]botrecord.Visit, error) {
	return store.visits, nil
}

// speechCounter records delivered bot chat at the extension boundary.
type speechCounter struct {
	// count stores intercepted message count.
	count atomic.Int32
}

// Intercept records and preserves one bot message.
func (counter *speechCounter) Intercept(_ context.Context, _ sdkbot.Bot, message string, _ sdkbot.Scope, _ int64) (string, bool, error) {
	counter.count.Add(1)
	return message, false, nil
}

// slowBehavior exposes asynchronous chat dispatch timing.
type slowBehavior struct {
	// started receives when slow work begins.
	started chan struct{}
}

// Type returns the test discriminator.
func (*slowBehavior) Type() string { return "slow" }

// OnPlace accepts placement.
func (*slowBehavior) OnPlace(context.Context, sdkbot.Bot, sdkbot.Actions) error { return nil }

// OnPickup accepts pickup.
func (*slowBehavior) OnPickup(context.Context, sdkbot.Bot, sdkbot.Actions) error { return nil }

// OnCycle accepts room cycles.
func (*slowBehavior) OnCycle(context.Context, sdkbot.Bot, sdkbot.Actions) error { return nil }

// OnUserSay deliberately blocks only a shared worker.
func (behavior *slowBehavior) OnUserSay(context.Context, sdkbot.Bot, sdkbot.Message, sdkbot.Actions) error {
	select {
	case behavior.started <- struct{}{}:
	default:
	}
	time.Sleep(200 * time.Millisecond)
	return nil
}

// OnUserEnter accepts room entries.
func (*slowBehavior) OnUserEnter(context.Context, sdkbot.Bot, int64, sdkbot.Actions) error {
	return nil
}

// SaveCustomSkill rejects unsupported test skills.
func (*slowBehavior) SaveCustomSkill(context.Context, sdkbot.Bot, int32, string) error {
	return sdkbot.ErrUnsupportedSkill
}

// TestSlowBehaviorDoesNotBlockRoomCycle verifies bounded shared-worker dispatch.
func TestSlowBehaviorDoesNotBlockRoomCycle(t *testing.T) {
	started := make(chan struct{}, 1)
	registry := botbehavior.NewRegistry()
	if err := registry.Register("slow", func() sdkbot.Behavior { return &slowBehavior{started: started} }); err != nil {
		t.Fatalf("register: %v", err)
	}
	service, room := serviceFixture(t, registry, []botrecord.Bot{placedBot(1, "slow", false)})
	service.Start()
	t.Cleanup(service.Stop)
	if err := service.EnsureRoom(context.Background(), room); err != nil {
		t.Fatalf("ensure room: %v", err)
	}
	begin := time.Now()
	service.HandleUserSay(context.Background(), room.ID(), 9, "hello")
	if elapsed := time.Since(begin); elapsed > 50*time.Millisecond {
		t.Fatalf("chat dispatch blocked for %s", elapsed)
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("slow behavior did not start")
	}
	begin = time.Now()
	if err := service.Cycle(context.Background(), room, time.Now()); err != nil {
		t.Fatalf("cycle: %v", err)
	}
	if elapsed := time.Since(begin); elapsed > 50*time.Millisecond {
		t.Fatalf("room cycle blocked for %s", elapsed)
	}
}

// TestBartenderMatchesWholeWordsWithinDistance verifies keyword boundaries.
func TestBartenderMatchesWholeWordsWithinDistance(t *testing.T) {
	registry := botbehavior.NewRegistry()
	if err := botbehavior.RegisterBuiltins(registry); err != nil {
		t.Fatalf("register: %v", err)
	}
	bot := placedBot(1, botrecord.BehaviorBartender, false)
	bot.X = integerPointer(8)
	bot.Y = integerPointer(2)
	store := &roomStore{bots: []botrecord.Bot{bot}, serveItems: []botrecord.ServeItem{{Keyword: "tea", DefinitionID: 5}}}
	service, room := serviceFixtureWithStore(t, registry, store)
	if err := service.EnsureRoom(context.Background(), room); err != nil {
		t.Fatalf("ensure: %v", err)
	}
	if _, err := room.Join(roomlive.Occupant{PlayerID: 9, ConnectionID: "9", ConnectionKind: "test"}); err != nil {
		t.Fatalf("join: %v", err)
	}
	view := sdkbot.Bot{ID: 1, OwnerPlayerID: 1, RoomID: 9, BehaviorType: botrecord.BehaviorBartender}
	matched, err := service.ServeKeyword(context.Background(), view, sdkbot.Message{PlayerID: 9, Text: "teapot"})
	if err != nil || matched {
		t.Fatalf("substring matched=%t err=%v", matched, err)
	}
	matched, err = service.ServeKeyword(context.Background(), view, sdkbot.Message{PlayerID: 9, Text: "tea please"})
	if err != nil || !matched {
		t.Fatalf("word matched=%t err=%v", matched, err)
	}
}

// TestVisitorLogPromptsAndListsOnlyOncePerPlacement verifies runtime guards.
func TestVisitorLogPromptsAndListsOnlyOncePerPlacement(t *testing.T) {
	registry := botbehavior.NewRegistry()
	if err := botbehavior.RegisterBuiltins(registry); err != nil {
		t.Fatalf("register: %v", err)
	}
	bot := placedBot(1, botrecord.BehaviorVisitorLog, false)
	bot.X = integerPointer(1)
	store := &roomStore{bots: []botrecord.Bot{bot}, visits: []botrecord.Visit{{PlayerID: 2, PlayerName: "Alice", EnteredAt: time.Now()}}}
	service, room := serviceFixtureWithStore(t, registry, store)
	counter := &speechCounter{}
	service.SetSpeechInterceptor(counter)
	service.Start()
	t.Cleanup(service.Stop)
	if _, err := room.Join(roomlive.Occupant{PlayerID: 1, ConnectionID: "1", ConnectionKind: "test"}); err != nil {
		t.Fatalf("join: %v", err)
	}
	for range 2 {
		if err := service.HandleUserEnter(context.Background(), room.ID(), 1); err != nil {
			t.Fatalf("enter: %v", err)
		}
	}
	waitCount(t, counter, 1)
	service.HandleUserSay(context.Background(), room.ID(), 1, "sí")
	service.HandleUserSay(context.Background(), room.ID(), 1, "sí")
	waitCount(t, counter, 2)
	time.Sleep(20 * time.Millisecond)
	if value := counter.count.Load(); value != 2 {
		t.Fatalf("expected two total messages, got %d", value)
	}
}

// serviceFixture creates a loaded active room and bot core.
func serviceFixture(t testing.TB, behaviors *botbehavior.Registry, bots []botrecord.Bot) (*botcore.Service, *roomlive.Room) {
	return serviceFixtureWithStore(t, behaviors, &roomStore{bots: bots})
}

// serviceFixtureWithStore creates a loaded bot core with a configurable store.
func serviceFixtureWithStore(t testing.TB, behaviors *botbehavior.Registry, store *roomStore) (*botcore.Service, *roomlive.Room) {
	t.Helper()
	rooms := roomlive.NewRegistry(nil)
	active, err := rooms.Activate(roomlive.Snapshot{ID: 9, OwnerPlayerID: 1, MaxUsers: 100})
	if err != nil {
		t.Fatalf("activate: %v", err)
	}
	t.Cleanup(func() { active.Close() })
	roomGrid, err := grid.Parse("0000000000\r0000000000\r0000000000", grid.WithDoor(9, 2))
	if err != nil {
		t.Fatalf("grid: %v", err)
	}
	if err = active.LoadWorld(roomlive.WorldConfig{Grid: roomGrid, Door: worldpath.Position{Point: grid.MustPoint(9, 2)}, Body: worldunit.RotationSouth, Head: worldunit.RotationSouth}); err != nil {
		t.Fatalf("load world: %v", err)
	}
	service := botcore.New(botpolicy.Config{}, store, rooms, nil, playerlive.NewRegistry(), nil, nil, behaviors, nil, nil, nil, nil, nil, zap.NewNop())
	return service, active
}

// placedBot creates a valid placed record.
func placedBot(id int64, behavior string, canWalk bool) botrecord.Bot {
	roomID, x, y, z, rotation := int64(9), int(id-1), 0, float64(0), int16(2)
	return botrecord.Bot{ID: id, OwnerPlayerID: 1, RoomID: &roomID, BehaviorType: behavior, Name: "Bot", Figure: "hd-180-1", Gender: "M", X: &x, Y: &y, Z: &z, Rotation: &rotation, CanWalk: canWalk, ChatDelaySeconds: 10}
}

// integerPointer returns one stable coordinate pointer.
func integerPointer(value int) *int { return &value }

// waitCount waits for asynchronous behavior delivery.
func waitCount(t testing.TB, counter *speechCounter, expected int32) {
	t.Helper()
	deadline := time.Now().Add(time.Second)
	for time.Now().Before(deadline) {
		if counter.count.Load() >= expected {
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatalf("expected count %d, got %d", expected, counter.count.Load())
}
