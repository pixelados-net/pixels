package command

import (
	"context"
	"errors"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

// testCommand is a named test command.
type testCommand struct {
	// name stores the command name.
	name Name
}

// CommandName returns the test command name.
func (command testCommand) CommandName() Name {
	return command.name
}

// TestDispatcherDispatchesEnvelope verifies valid commands reach handlers.
func TestDispatcherDispatchesEnvelope(t *testing.T) {
	var handled Envelope[testCommand]
	dispatcher := mustDispatcher(t, HandlerFunc[testCommand](func(ctx context.Context, envelope Envelope[testCommand]) error {
		handled = envelope
		return nil
	}))

	envelope := Envelope[testCommand]{
		Command:  testCommand{name: "player.enter_room"},
		Metadata: Metadata{PlayerID: 10, ConnectionID: "ws-1"},
	}

	if err := dispatcher.Dispatch(context.Background(), envelope); err != nil {
		t.Fatalf("dispatch command: %v", err)
	}
	if handled.Command.CommandName() != "player.enter_room" {
		t.Fatalf("expected command to be handled, got %s", handled.Command.CommandName())
	}
	if handled.Metadata.CreatedAt.IsZero() {
		t.Fatal("expected dispatcher to fill creation time")
	}
}

// TestDispatcherRejectsInvalidCommand verifies unnamed commands are rejected.
func TestDispatcherRejectsInvalidCommand(t *testing.T) {
	dispatcher := mustDispatcher(t, HandlerFunc[testCommand](func(context.Context, Envelope[testCommand]) error {
		return nil
	}))

	err := dispatcher.Dispatch(context.Background(), Envelope[testCommand]{Command: testCommand{}})
	if !errors.Is(err, ErrInvalidName) {
		t.Fatalf("expected invalid name error, got %v", err)
	}
}

// TestNewDispatcherRejectsInvalidHandler verifies nil handlers are rejected.
func TestNewDispatcherRejectsInvalidHandler(t *testing.T) {
	var handler Handler[testCommand]
	dispatcher, err := NewDispatcher(handler)
	if !errors.Is(err, ErrInvalidHandler) {
		t.Fatalf("expected invalid handler error, got %v", err)
	}
	if dispatcher != nil {
		t.Fatal("expected nil dispatcher")
	}
}

// TestChainAppliesMiddlewareOrder verifies middleware wraps in declaration order.
func TestChainAppliesMiddlewareOrder(t *testing.T) {
	order := make([]string, 0, 3)
	handler := HandlerFunc[testCommand](func(context.Context, Envelope[testCommand]) error {
		order = append(order, "handler")
		return nil
	})
	first := func(next Handler[testCommand]) Handler[testCommand] {
		return HandlerFunc[testCommand](func(ctx context.Context, envelope Envelope[testCommand]) error {
			order = append(order, "first")
			return next.Handle(ctx, envelope)
		})
	}
	second := func(next Handler[testCommand]) Handler[testCommand] {
		return HandlerFunc[testCommand](func(ctx context.Context, envelope Envelope[testCommand]) error {
			order = append(order, "second")
			return next.Handle(ctx, envelope)
		})
	}

	dispatcher, err := NewDispatcher(handler, first, second)
	if err != nil {
		t.Fatalf("create dispatcher: %v", err)
	}
	err = dispatcher.Dispatch(context.Background(), Envelope[testCommand]{Command: testCommand{name: "test.command"}})
	if err != nil {
		t.Fatalf("dispatch command: %v", err)
	}

	assertCommandOrder(t, order, []string{"first", "second", "handler"})
}

// TestDispatcherLogsDebugCommand verifies valid commands emit debug logs.
func TestDispatcherLogsDebugCommand(t *testing.T) {
	core, logs := observer.New(zap.DebugLevel)
	dispatcher := mustDispatcher(t, HandlerFunc[testCommand](func(context.Context, Envelope[testCommand]) error {
		return nil
	})).WithLogger(zap.New(core))

	envelope := Envelope[testCommand]{
		Command:  testCommand{name: "player.walk"},
		Metadata: Metadata{PlayerID: 7, ConnectionID: "ws-7"},
	}
	if err := dispatcher.Dispatch(context.Background(), envelope); err != nil {
		t.Fatalf("dispatch command: %v", err)
	}

	entries := logs.FilterMessage("command dispatched").All()
	if len(entries) != 1 {
		t.Fatalf("expected one debug log, got %d", len(entries))
	}
	if entries[0].ContextMap()["command_name"] != "player.walk" {
		t.Fatalf("expected command name field, got %#v", entries[0].ContextMap())
	}
}

// mustDispatcher creates a test dispatcher.
func mustDispatcher(t *testing.T, handler Handler[testCommand]) *Dispatcher[testCommand] {
	t.Helper()

	dispatcher, err := NewDispatcher(handler)
	if err != nil {
		t.Fatalf("create dispatcher: %v", err)
	}

	return dispatcher
}

// assertCommandOrder verifies middleware execution order.
func assertCommandOrder(t *testing.T, actual []string, expected []string) {
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
