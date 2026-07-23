package behavior

import (
	"errors"
	"testing"

	sdkbot "github.com/niflaot/pixels/sdk/bot"
)

// TestRegistryRejectsDuplicatesAndFallsBack verifies extension safety.
func TestRegistryRejectsDuplicatesAndFallsBack(t *testing.T) {
	registry := NewRegistry()
	if err := registry.Register("generic", func() sdkbot.Behavior { return nil }); err == nil {
		t.Fatal("expected invalid factory")
	}
	if err := registry.Register("generic", func() sdkbot.Behavior { return Generic{} }); err != nil {
		t.Fatalf("register generic: %v", err)
	}
	if err := registry.Register("generic", func() sdkbot.Behavior { return Generic{} }); !errors.Is(err, ErrAlreadyRegistered) {
		t.Fatalf("duplicate error=%v", err)
	}
	if resolved := registry.For("removed-behavior"); resolved == nil || resolved.Type() != "generic" {
		t.Fatalf("unexpected fallback %#v", resolved)
	}
}
