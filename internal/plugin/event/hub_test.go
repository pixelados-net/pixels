package event

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	pluginruntime "github.com/niflaot/pixels/internal/plugin/runtime"
	playerconnected "github.com/niflaot/pixels/internal/realm/player/events/connected"
	"github.com/niflaot/pixels/pkg/bus"
	sdkevent "github.com/niflaot/pixels/sdk/event"
	sdkplayer "github.com/niflaot/pixels/sdk/player"
	"go.uber.org/zap"
)

// testPlayer creates one immutable connected-player fixture.
func testPlayer() sdkplayer.Player {
	return sdkplayer.Player{ID: 7, Username: "demo", RoomID: 3, Online: true}
}

// playerFinder stores one optional forwarded-event player.
type playerFinder struct {
	// found controls whether the player is online.
	found bool
}

// Find returns the fixture connected player.
func (finder playerFinder) Find(int64) (sdkplayer.Player, bool) {
	return testPlayer(), finder.found
}

// TestHubOrdersAndCancelsListeners verifies priority and IgnoreCancelled semantics.
func TestHubOrdersAndCancelsListeners(t *testing.T) {
	hub := NewHub(time.Second, zap.NewNop())
	event := sdkevent.NewChatSend(testPlayer(), 7, "hello")
	order := make([]string, 0, 2)
	if err := hub.listen(pluginruntime.NewScope("high"), event.Name(), sdkevent.ListenerOptions{Priority: 100}, func(_ context.Context, current sdkevent.Event) error {
		order = append(order, "high")
		current.(sdkevent.Cancellable).SetCancelled(true)
		return nil
	}); err != nil {
		t.Fatalf("listen high: %v", err)
	}
	_ = hub.listen(pluginruntime.NewScope("skipped"), event.Name(), sdkevent.ListenerOptions{IgnoreCancelled: true}, func(context.Context, sdkevent.Event) error {
		order = append(order, "skipped")
		return nil
	})
	_ = hub.listen(pluginruntime.NewScope("low"), event.Name(), sdkevent.ListenerOptions{Priority: -100}, func(context.Context, sdkevent.Event) error {
		order = append(order, "low")
		return nil
	})
	err := hub.Dispatch(context.Background(), event)
	if !errors.Is(err, ErrEventCancelled) || !reflect.DeepEqual(order, []string{"high", "low"}) {
		t.Fatalf("expected ordered cancellation, order=%v err=%v", order, err)
	}
}

// TestHubIsolatesPanickingPlugin verifies one panic does not block other scopes.
func TestHubIsolatesPanickingPlugin(t *testing.T) {
	hub := NewHub(time.Second, zap.NewNop())
	broken := pluginruntime.NewScope("broken")
	called := false
	_ = hub.listen(broken, sdkevent.PlayerConnectedName, sdkevent.ListenerOptions{Priority: 100}, func(context.Context, sdkevent.Event) error { panic("boom") })
	_ = hub.listen(pluginruntime.NewScope("healthy"), sdkevent.PlayerConnectedName, sdkevent.ListenerOptions{}, func(context.Context, sdkevent.Event) error {
		called = true
		return nil
	})
	err := hub.Dispatch(context.Background(), &sdkevent.PlayerConnected{Player: testPlayer()})
	if err != nil || broken.Enabled() || !called {
		t.Fatalf("expected isolated panic, called=%v enabled=%v err=%v", called, broken.Enabled(), err)
	}
}

// TestHubBoundsListener verifies stalled callbacks return within their deadline.
func TestHubBoundsListener(t *testing.T) {
	hub := NewHub(time.Millisecond, zap.NewNop())
	_ = hub.listen(pluginruntime.NewScope("slow"), sdkevent.PlayerConnectedName, sdkevent.ListenerOptions{}, func(ctx context.Context, _ sdkevent.Event) error {
		<-ctx.Done()
		return ctx.Err()
	})
	if err := hub.Dispatch(context.Background(), &sdkevent.PlayerConnected{Player: testPlayer()}); err != nil {
		t.Fatalf("dispatch timed listener: %v", err)
	}
}

// TestHubAccessDispatchChatAndValidation verifies facade registration and mutation.
func TestHubAccessDispatchChatAndValidation(t *testing.T) {
	hub := NewHub(time.Second, zap.NewNop())
	access := NewAccess(hub, pluginruntime.NewScope("rewrite"))
	if err := access.Listen(sdkevent.ChatSendName, sdkevent.ListenerOptions{}, func(_ context.Context, current sdkevent.Event) error {
		chat := current.(*sdkevent.ChatSend)
		chat.Text = "changed"
		chat.SetCancelled(true)
		return nil
	}); err != nil {
		t.Fatalf("listen through access: %v", err)
	}
	text, cancelled := hub.DispatchChat(context.Background(), testPlayer(), 3, "original")
	if text != "changed" || !cancelled {
		t.Fatalf("expected changed cancellation, text=%q cancelled=%v", text, cancelled)
	}
	if err := hub.listen(nil, "", sdkevent.ListenerOptions{}, nil); !errors.Is(err, ErrInvalidListener) {
		t.Fatalf("expected invalid listener, got %v", err)
	}
	if err := hub.Dispatch(context.Background(), nil); !errors.Is(err, ErrInvalidListener) {
		t.Fatalf("expected invalid event, got %v", err)
	}
}

// TestHubForwardsPlayerConnected verifies the internal bus remains read-only behind the SDK.
func TestHubForwardsPlayerConnected(t *testing.T) {
	local := bus.New()
	hub := NewHub(time.Second, zap.NewNop())
	if err := hub.RegisterPlayerConnected(local, playerFinder{found: true}); err != nil {
		t.Fatalf("register forwarding: %v", err)
	}
	called := false
	_ = hub.listen(pluginruntime.NewScope("listener"), sdkevent.PlayerConnectedName, sdkevent.ListenerOptions{}, func(context.Context, sdkevent.Event) error {
		called = true
		return nil
	})
	if err := local.Publish(context.Background(), bus.Event{Name: playerconnected.Name, Payload: playerconnected.Payload{PlayerID: 7}}); err != nil {
		t.Fatalf("publish connected: %v", err)
	}
	if !called {
		t.Fatal("expected forwarded connected event")
	}
	if err := local.Publish(context.Background(), bus.Event{Name: playerconnected.Name, Payload: "invalid"}); err != nil {
		t.Fatalf("publish invalid payload: %v", err)
	}
}

// TestHubDiscardsLateTimedOutMutation verifies mutable events never escape their callback.
func TestHubDiscardsLateTimedOutMutation(t *testing.T) {
	hub := NewHub(time.Millisecond, zap.NewNop())
	release := make(chan struct{})
	done := make(chan struct{})
	_ = hub.listen(pluginruntime.NewScope("slow"), sdkevent.ChatSendName, sdkevent.ListenerOptions{}, func(_ context.Context, current sdkevent.Event) error {
		<-release
		current.(*sdkevent.ChatSend).Text = "late"
		close(done)
		return nil
	})
	event := sdkevent.NewChatSend(testPlayer(), 3, "original")
	if err := hub.Dispatch(context.Background(), event); err != nil {
		t.Fatalf("dispatch timed listener: %v", err)
	}
	close(release)
	<-done
	if event.Text != "original" {
		t.Fatalf("expected original event after timeout, got %q", event.Text)
	}
}
