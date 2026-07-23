package bridge

import (
	"context"
	"testing"
	"time"

	roomlive "github.com/niflaot/pixels/internal/realm/room/runtime/live"
	roomwired "github.com/niflaot/pixels/internal/realm/room/world/wired"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/configuration"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/effect"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/record"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
	wiredruntime "github.com/niflaot/pixels/internal/realm/room/world/wired/runtime"
	"github.com/niflaot/pixels/internal/realm/room/world/wired/trigger"
	sdkbot "github.com/niflaot/pixels/sdk/bot"
)

// speechStore returns one immutable speech stack for every requested test room.
type speechStore struct{}

// LoadRoom returns one SAY trigger and message effect stack.
func (speechStore) LoadRoom(_ context.Context, roomID int64) ([]record.Config, error) {
	return []record.Config{
		{ItemID: 1, RoomID: roomID, Interaction: "wf_trg_says_something", X: 1, Y: 1, StringParam: "hello"},
		{ItemID: 2, RoomID: roomID, Interaction: "wf_act_show_message", X: 1, Y: 1, StringParam: "received"},
	}, nil
}

// Find is unused by speech bridge tests.
func (speechStore) Find(context.Context, int64, int64) (record.Config, bool, error) {
	return record.Config{}, false, nil
}

// Save is unused by speech bridge tests.
func (speechStore) Save(context.Context, record.Config, int64) (record.Config, error) {
	return record.Config{}, nil
}

// Capture is unused by speech bridge tests.
func (speechStore) Capture(context.Context, int64, int64) ([]record.Target, error) { return nil, nil }

// SaveRewardConfig is unused by speech bridge tests.
func (speechStore) SaveRewardConfig(context.Context, record.Config, int64, []record.Reward) (record.Config, error) {
	return record.Config{}, nil
}

// CleanupItem is unused by speech bridge tests.
func (speechStore) CleanupItem(context.Context, int64) error { return nil }

// speechAvatar observes asynchronously executed SAY stacks.
type speechAvatar struct {
	// called reports one executed message effect.
	called chan struct{}
}

// ExecuteAvatar observes one player-facing effect.
func (avatar *speechAvatar) ExecuteAvatar(context.Context, effect.AvatarOperation, *configuration.Node, trigger.Event) (effect.Result, error) {
	select {
	case avatar.called <- struct{}{}:
	default:
	}
	return effect.Result{Status: effect.Applied}, nil
}

// TestSpeechBridgeSchedulesMatchingPlayerAndBotSpeech verifies filtered speech reaches the room owner loop.
func TestSpeechBridgeSchedulesMatchingPlayerAndBotSpeech(t *testing.T) {
	rooms := roomlive.NewRegistry(nil)
	active, err := rooms.Activate(roomlive.Snapshot{ID: 1, MaxUsers: 5})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _, _ = rooms.Close(context.Background(), active.ID()) })
	avatar := &speechAvatar{called: make(chan struct{}, 2)}
	registered, err := registry.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	compiler := configuration.NewCompiler(registered, roomwired.Config{})
	engine := wiredruntime.New(roomwired.Config{Enabled: true}, speechStore{}, compiler, effect.New(effect.Services{Avatar: avatar}), nil, nil, nil)
	if err = engine.Reload(context.Background(), 1, time.Now()); err != nil {
		t.Fatal(err)
	}
	bridge := NewSpeechBridge(rooms, engine)
	consumed, err := bridge.Intercept(context.Background(), 1, 7, "demo", "well hello there")
	if err != nil || !consumed {
		t.Fatalf("player consumed=%t err=%v", consumed, err)
	}
	active.RunScheduled(time.Now().Add(time.Second))
	select {
	case <-avatar.called:
	case <-time.After(time.Second):
		t.Fatal("scheduled player speech did not execute")
	}
	interceptor := NewBotSpeechInterceptor(bridge)
	message, consumed, err := interceptor.Intercept(context.Background(), sdkbot.Bot{ID: 5, RoomID: 1, Name: "WiredGuide"}, "hello", sdkbot.ScopeTalk, 0)
	if err != nil || !consumed || message != "hello" {
		t.Fatalf("bot message=%q consumed=%t err=%v", message, consumed, err)
	}
	active.RunScheduled(time.Now().Add(time.Second))
	consumed, err = bridge.InterceptPet(context.Background(), 1, -99, "hello")
	if err != nil || !consumed {
		t.Fatalf("pet consumed=%t err=%v", consumed, err)
	}
	active.RunScheduled(time.Now().Add(time.Second))
}

// TestSpeechBridgeSkipsUnmatchedMissingAndRecursiveSpeech verifies all recursion guard exits.
func TestSpeechBridgeSkipsUnmatchedMissingAndRecursiveSpeech(t *testing.T) {
	rooms := roomlive.NewRegistry(nil)
	registered, err := registry.Canonical()
	if err != nil {
		t.Fatal(err)
	}
	engine := wiredruntime.New(roomwired.Config{Enabled: true}, speechStore{}, configuration.NewCompiler(registered, roomwired.Config{}), effect.New(effect.Services{}), nil, nil, nil)
	if err = engine.Reload(context.Background(), 1, time.Now()); err != nil {
		t.Fatal(err)
	}
	bridge := NewSpeechBridge(rooms, engine)
	if consumed, err := bridge.Intercept(context.Background(), 1, 7, "demo", "other"); err != nil || consumed {
		t.Fatalf("unmatched consumed=%t err=%v", consumed, err)
	}
	if consumed, err := bridge.Intercept(context.Background(), 1, 7, "demo", "hello"); err != nil || consumed {
		t.Fatalf("missing room consumed=%t err=%v", consumed, err)
	}
	ctx := context.WithValue(context.Background(), speechDepthKey{}, 4)
	if consumed, err := bridge.Intercept(ctx, 1, 7, "demo", "hello"); err != nil || consumed {
		t.Fatalf("recursive consumed=%t err=%v", consumed, err)
	}
}
