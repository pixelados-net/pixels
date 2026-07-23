package connection

import (
	"context"
	"errors"
	"sync"
	"testing"
)

// TestRegistryRegisterGetRemove verifies kind-scoped registration.
func TestRegistryRegisterGetRemove(t *testing.T) {
	registry := NewRegistry()
	websocket := mustSession(t, sessionFixture(t))
	raw := mustSession(t, sessionFixtureWithKind(t, "raw"))

	if err := registry.Register(websocket); err != nil {
		t.Fatalf("register websocket: %v", err)
	}

	if err := registry.Register(raw); err != nil {
		t.Fatalf("register raw: %v", err)
	}

	if registry.CountAll() != 2 {
		t.Fatalf("expected %d connections, got %d", 2, registry.CountAll())
	}

	if _, exists := registry.Get("websocket", "one"); !exists {
		t.Fatal("expected websocket connection")
	}

	if _, exists := registry.Get("raw", "one"); !exists {
		t.Fatal("expected raw connection")
	}

	if _, removed := registry.Remove("websocket", "one"); !removed {
		t.Fatal("expected removed websocket")
	}
}

// TestRegistryListAll verifies registry-wide snapshots.
func TestRegistryListAll(t *testing.T) {
	registry := NewRegistry()
	mustRegisterConnection(t, registry, mustSession(t, sessionFixture(t)))
	mustRegisterConnection(t, registry, mustSession(t, sessionFixtureWithKind(t, "raw")))

	connections := registry.ListAll()
	if len(connections) != 2 {
		t.Fatalf("expected two connections, got %d", len(connections))
	}
}

// TestRegistryRejectsDuplicates verifies duplicate connection protection.
func TestRegistryRejectsDuplicates(t *testing.T) {
	registry := NewRegistry()
	session := mustSession(t, sessionFixture(t))

	if err := registry.Register(session); err != nil {
		t.Fatalf("register session: %v", err)
	}

	if err := registry.Register(session); !errors.Is(err, ErrConnectionExists) {
		t.Fatalf("expected connection exists, got %v", err)
	}
}

// TestRegistryDisconnect verifies removal and disposal.
func TestRegistryDisconnect(t *testing.T) {
	registry := NewRegistry()
	session := mustSession(t, sessionFixture(t))

	if err := registry.Register(session); err != nil {
		t.Fatalf("register session: %v", err)
	}

	if err := registry.Disconnect(context.Background(), "websocket", "one", UnknownReason()); err != nil {
		t.Fatalf("disconnect session: %v", err)
	}

	if registry.Count("websocket") != 0 {
		t.Fatalf("expected empty registry, got %d", registry.Count("websocket"))
	}
}

// TestRegistryDisconnectGroups verifies group and all disconnection.
func TestRegistryDisconnectGroups(t *testing.T) {
	registry := NewRegistry()
	mustRegisterConnection(t, registry, mustSession(t, sessionFixture(t)))
	mustRegisterConnection(t, registry, mustSession(t, sessionFixtureWithID(t, "two")))
	mustRegisterConnection(t, registry, mustSession(t, sessionFixtureWithKind(t, "raw")))

	if errors := registry.DisconnectKind(context.Background(), "websocket", UnknownReason()); len(errors) != 0 {
		t.Fatalf("expected no errors, got %d", len(errors))
	}

	if registry.Count("websocket") != 0 {
		t.Fatalf("expected no websocket connections, got %d", registry.Count("websocket"))
	}

	if errors := registry.DisconnectAll(context.Background(), UnknownReason()); len(errors) != 0 {
		t.Fatalf("expected no errors, got %d", len(errors))
	}

	if registry.CountAll() != 0 {
		t.Fatalf("expected empty registry, got %d", registry.CountAll())
	}
}

// TestRegistryConcurrentRegister verifies mutex-safe registration.
func TestRegistryConcurrentRegister(t *testing.T) {
	registry := NewRegistry()
	wait := sync.WaitGroup{}

	for index := 0; index < 20; index++ {
		wait.Add(1)
		go func(index int) {
			defer wait.Done()
			session := mustSession(t, sessionFixtureWithID(t, ID(rune('a'+index))))
			if err := registry.Register(session); err != nil {
				t.Errorf("register session: %v", err)
			}
		}(index)
	}

	wait.Wait()

	if registry.Count("websocket") != 20 {
		t.Fatalf("expected %d connections, got %d", 20, registry.Count("websocket"))
	}
}

// TestRegistryMissingDisconnect verifies missing connection errors.
func TestRegistryMissingDisconnect(t *testing.T) {
	registry := NewRegistry()
	err := registry.Disconnect(context.Background(), "websocket", "missing", UnknownReason())
	if !errors.Is(err, ErrConnectionNotFound) {
		t.Fatalf("expected missing connection, got %v", err)
	}
}

// sessionFixtureWithID returns a session fixture with an id override.
func sessionFixtureWithID(t *testing.T, id ID) sessionFixtureConfig {
	t.Helper()
	fixture := sessionFixture(t)
	fixture.ID = id

	return fixture
}

// sessionFixtureWithKind returns a session fixture with a kind override.
func sessionFixtureWithKind(t *testing.T, kind Kind) sessionFixtureConfig {
	t.Helper()
	fixture := sessionFixture(t)
	fixture.Kind = kind

	return fixture
}

// mustRegisterConnection registers a connection or fails the test.
func mustRegisterConnection(t *testing.T, registry *Registry, connection Connection) {
	t.Helper()
	if err := registry.Register(connection); err != nil {
		t.Fatalf("register connection: %v", err)
	}
}
