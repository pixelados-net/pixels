package binding

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/networking/connection"
)

// TestRegistryAddFindAndRemove verifies binding registry lifecycle.
func TestRegistryAddFindAndRemove(t *testing.T) {
	registry := NewRegistry()
	value := testBinding(10, "ws-1")

	if err := registry.Add(value); err != nil {
		t.Fatalf("add binding: %v", err)
	}
	if registry.Count() != 1 {
		t.Fatalf("expected count 1, got %d", registry.Count())
	}

	byPlayer, found := registry.FindByPlayer(10)
	if !found || byPlayer.ConnectionID != "ws-1" {
		t.Fatalf("expected player binding, got %#v", byPlayer)
	}

	byConnection, found := registry.FindByConnection(ConnectionKey{Kind: "websocket", ID: "ws-1"})
	if !found || byConnection.PlayerID != 10 {
		t.Fatalf("expected connection binding, got %#v", byConnection)
	}

	removed, found := registry.RemoveByConnection(ConnectionKey{Kind: "websocket", ID: "ws-1"})
	if !found || removed.PlayerID != 10 {
		t.Fatalf("expected removed binding, got %#v", removed)
	}
	if registry.Count() != 0 {
		t.Fatalf("expected empty registry, got %d", registry.Count())
	}
}

// TestRegistryRejectsInvalidBinding verifies validation.
func TestRegistryRejectsInvalidBinding(t *testing.T) {
	registry := NewRegistry()

	err := registry.Add(Binding{})
	if !errors.Is(err, ErrInvalidBinding) {
		t.Fatalf("expected invalid binding error, got %v", err)
	}
}

// TestRegistryRejectsDuplicates verifies player and connection uniqueness.
func TestRegistryRejectsDuplicates(t *testing.T) {
	registry := NewRegistry()
	if err := registry.Add(testBinding(10, "ws-1")); err != nil {
		t.Fatalf("add binding: %v", err)
	}

	if err := registry.Add(testBinding(10, "ws-2")); !errors.Is(err, ErrBindingExists) {
		t.Fatalf("expected duplicate player error, got %v", err)
	}
	if err := registry.Add(testBinding(11, "ws-1")); !errors.Is(err, ErrBindingExists) {
		t.Fatalf("expected duplicate connection error, got %v", err)
	}
}

// TestRegistrySnapshotCopiesBindings verifies snapshots are detached.
func TestRegistrySnapshotCopiesBindings(t *testing.T) {
	registry := NewRegistry()
	if err := registry.Add(testBinding(10, "ws-1")); err != nil {
		t.Fatalf("add binding: %v", err)
	}

	snapshot := registry.Snapshot()
	if len(snapshot) != 1 {
		t.Fatalf("expected one binding, got %d", len(snapshot))
	}

	registry.RemoveByPlayer(10)
	if len(snapshot) != 1 {
		t.Fatalf("expected snapshot to stay stable, got %d", len(snapshot))
	}
}

// testBinding creates a registry test binding.
func testBinding(playerID int64, connectionID connection.ID) Binding {
	return Binding{PlayerID: playerID, ConnectionID: connectionID, ConnectionKind: "websocket"}
}
