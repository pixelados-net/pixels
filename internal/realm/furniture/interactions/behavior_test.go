package interactions

import "testing"

// TestRegistryResolve verifies supported behavior routing and nil safety.
func TestRegistryResolve(t *testing.T) {
	registry := NewRegistry()
	for _, interactionType := range []string{"default", "toggle", "gate"} {
		if behavior, found := registry.Resolve(interactionType); !found || behavior == nil {
			t.Fatalf("expected behavior for %q", interactionType)
		}
	}
	if behavior, found := registry.Resolve("teleport"); found || behavior != nil {
		t.Fatalf("expected unsupported behavior, got %#v", behavior)
	}
	var missing *Registry
	if behavior, found := missing.Resolve("gate"); found || behavior != nil {
		t.Fatalf("expected nil registry miss, got %#v", behavior)
	}
}
