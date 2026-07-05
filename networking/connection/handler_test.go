package connection

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/codec"
)

// TestHandlerRegistryHandle verifies registered handler routing.
func TestHandlerRegistryHandle(t *testing.T) {
	registry := NewHandlerRegistry()
	called := false
	handler := func(context Context, packet codec.Packet) error {
		called = context.ConnectionID == "one" && packet.Header == 7
		return nil
	}

	if err := registry.Register(7, handler, AllowAnyActiveState(), AllowUnauthenticated()); err != nil {
		t.Fatalf("register handler: %v", err)
	}

	if err := registry.Handle(Context{ConnectionID: "one", State: StateCreated}, codec.Packet{Header: 7}); err != nil {
		t.Fatalf("handle packet: %v", err)
	}

	if !called {
		t.Fatal("expected handler call")
	}
}

// TestHandlerRegistryRejectsDuplicate verifies duplicate handler protection.
func TestHandlerRegistryRejectsDuplicate(t *testing.T) {
	registry := NewHandlerRegistry()
	handler := func(Context, codec.Packet) error {
		return nil
	}

	if err := registry.Register(7, handler); err != nil {
		t.Fatalf("register handler: %v", err)
	}

	if err := registry.Register(7, handler); !errors.Is(err, ErrHandlerExists) {
		t.Fatalf("expected handler exists, got %v", err)
	}
}

// TestHandlerRegistryFallback verifies fallback packet handling.
func TestHandlerRegistryFallback(t *testing.T) {
	registry := NewHandlerRegistry()
	called := false
	registry.SetFallback(func(context Context, packet codec.Packet) error {
		called = packet.Header == 99
		return nil
	}, AllowAnyActiveState(), AllowUnauthenticated())

	if err := registry.Handle(Context{State: StateCreated}, codec.Packet{Header: 99}); err != nil {
		t.Fatalf("handle fallback: %v", err)
	}

	if !called {
		t.Fatal("expected fallback handler call")
	}
}

// TestHandlerRegistryMissing verifies missing handler errors.
func TestHandlerRegistryMissing(t *testing.T) {
	registry := NewHandlerRegistry()
	err := registry.Handle(Context{}, codec.Packet{Header: 99})
	if !errors.Is(err, ErrHandlerNotFound) {
		t.Fatalf("expected handler missing, got %v", err)
	}
}

// TestHandlerRegistryRejectsPolicy verifies default connected authentication gates.
func TestHandlerRegistryRejectsPolicy(t *testing.T) {
	registry := NewHandlerRegistry()
	if err := registry.Register(7, func(Context, codec.Packet) error { return nil }); err != nil {
		t.Fatalf("register handler: %v", err)
	}

	err := registry.Handle(Context{State: StateHandshaking}, codec.Packet{Header: 7})
	if !errors.Is(err, ErrHandlerPolicy) {
		t.Fatalf("expected handler policy error, got %v", err)
	}

	err = registry.Handle(Context{State: StateConnected, Authenticated: true}, codec.Packet{Header: 7})
	if err != nil {
		t.Fatalf("expected connected handler, got %v", err)
	}
}

// TestHandlerRegistryUnregister verifies handler removal.
func TestHandlerRegistryUnregister(t *testing.T) {
	registry := NewHandlerRegistry()
	handler := func(Context, codec.Packet) error {
		return nil
	}

	if err := registry.Register(7, handler); err != nil {
		t.Fatalf("register handler: %v", err)
	}

	if registry.Len() != 1 {
		t.Fatalf("expected %d handlers, got %d", 1, registry.Len())
	}

	if !registry.Unregister(7) {
		t.Fatal("expected handler removal")
	}

	if registry.Unregister(7) {
		t.Fatal("expected missing handler removal")
	}
}
