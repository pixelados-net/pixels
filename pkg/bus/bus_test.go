package bus

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

// TestBusPublishUsesPriority verifies higher priority listeners run first.
func TestBusPublishUsesPriority(t *testing.T) {
	local := New()
	defer func() {
		if err := local.Close(); err != nil {
			t.Fatalf("close bus: %v", err)
		}
	}()

	order := make([]string, 0, 3)
	mustSubscribe(t, local, "player.authenticated", PriorityLow, func(context.Context, Event) error {
		order = append(order, "low")
		return nil
	})
	mustSubscribe(t, local, "player.authenticated", PriorityHigh, func(context.Context, Event) error {
		order = append(order, "high")
		return nil
	})
	mustSubscribe(t, local, "player.authenticated", PriorityNormal, func(context.Context, Event) error {
		order = append(order, "normal")
		return nil
	})

	err := local.Publish(context.Background(), Event{Name: "player.authenticated", Payload: int64(1)})
	if err != nil {
		t.Fatalf("publish event: %v", err)
	}

	assertOrder(t, order, []string{"high", "normal", "low"})
}

// TestBusSubscribeOnceRunsOnce verifies one-shot listeners are removed.
func TestBusSubscribeOnceRunsOnce(t *testing.T) {
	local := New()
	defer func() {
		if err := local.Close(); err != nil {
			t.Fatalf("close bus: %v", err)
		}
	}()

	count := 0
	_, err := local.SubscribeOnce("player.disconnected", PriorityNormal, func(context.Context, Event) error {
		count++
		return nil
	})
	if err != nil {
		t.Fatalf("subscribe once: %v", err)
	}

	for range 2 {
		if err := local.Publish(context.Background(), Event{Name: "player.disconnected"}); err != nil {
			t.Fatalf("publish event: %v", err)
		}
	}

	if count != 1 {
		t.Fatalf("expected listener to run once, got %d", count)
	}
}

// TestBusUnsubscribeRemovesListener verifies removed listeners are skipped.
func TestBusUnsubscribeRemovesListener(t *testing.T) {
	local := New()
	defer func() {
		if err := local.Close(); err != nil {
			t.Fatalf("close bus: %v", err)
		}
	}()

	count := 0
	subscription := mustSubscribe(t, local, "catalog.page_opened", PriorityNormal, func(context.Context, Event) error {
		count++
		return nil
	})

	subscription.Unsubscribe()
	subscription.Unsubscribe()

	if err := local.Publish(context.Background(), Event{Name: "catalog.page_opened"}); err != nil {
		t.Fatalf("publish event: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected removed listener to be skipped, got %d", count)
	}
}

// TestBusPublishReturnsHandlerError verifies listener errors are returned.
func TestBusPublishReturnsHandlerError(t *testing.T) {
	local := New()
	defer func() {
		if err := local.Close(); err != nil {
			t.Fatalf("close bus: %v", err)
		}
	}()

	expected := errors.New("handler failed")
	mustSubscribe(t, local, "room.user_moved", PriorityNormal, func(context.Context, Event) error {
		return expected
	})

	err := local.Publish(context.Background(), Event{Name: "room.user_moved"})
	if !errors.Is(err, expected) {
		t.Fatalf("expected handler error, got %v", err)
	}
}

// TestBusPublishLogsDebugEvent verifies published events are observable in debug logs.
func TestBusPublishLogsDebugEvent(t *testing.T) {
	core, logs := observer.New(zap.DebugLevel)
	local := NewWithLogger(zap.New(core))
	defer func() {
		if err := local.Close(); err != nil {
			t.Fatalf("close bus: %v", err)
		}
	}()

	if err := local.Publish(context.Background(), Event{Name: "player.connected", Payload: int64(2)}); err != nil {
		t.Fatalf("publish event: %v", err)
	}

	entries := logs.FilterMessage("event published").All()
	if len(entries) != 1 {
		t.Fatalf("expected one debug log, got %d", len(entries))
	}
	if entries[0].ContextMap()["event_name"] != "player.connected" {
		t.Fatalf("expected event name field, got %#v", entries[0].ContextMap())
	}
}

// TestBusRejectsInvalidInput verifies subscription and publish validation.
func TestBusRejectsInvalidInput(t *testing.T) {
	local := New()

	if _, err := local.Subscribe("", PriorityNormal, func(context.Context, Event) error { return nil }); !errors.Is(err, ErrInvalidEvent) {
		t.Fatalf("expected invalid event error, got %v", err)
	}
	if _, err := local.Subscribe("player.authenticated", PriorityNormal, nil); !errors.Is(err, ErrInvalidHandler) {
		t.Fatalf("expected invalid handler error, got %v", err)
	}
	if err := local.Publish(context.Background(), Event{}); !errors.Is(err, ErrInvalidEvent) {
		t.Fatalf("expected invalid event error, got %v", err)
	}
}

// mustSubscribe registers a test subscription.
func mustSubscribe(t *testing.T, local *Bus, name Name, priority int, handler Handler) *Subscription {
	t.Helper()

	subscription, err := local.Subscribe(name, priority, handler)
	if err != nil {
		t.Fatalf("subscribe %s: %v", name, err)
	}

	return subscription
}

// assertOrder verifies listener execution order.
func assertOrder(t *testing.T, actual []string, expected []string) {
	t.Helper()

	if len(actual) != len(expected) {
		t.Fatalf("expected %v, got %v", expected, actual)
	}
	for index, value := range expected {
		if actual[index] != value {
			t.Fatalf("expected %v, got %v", expected, actual)
		}
	}
}
