// Package tests verifies the public WIRED registry contract.
package tests

import (
	"errors"
	"testing"

	"github.com/niflaot/pixels/internal/realm/room/world/wired/registry"
)

// TestCanonicalInventory verifies the audited family totals and aliases.
func TestCanonicalInventory(t *testing.T) {
	manifest := registry.CanonicalManifest()
	counts := map[registry.Family]int{}
	for _, descriptor := range manifest {
		counts[descriptor.Family]++
	}
	if len(manifest) != 76 || counts[registry.FamilyTrigger] != 17 || counts[registry.FamilyEffect] != 30 || counts[registry.FamilyCondition] != 24 || counts[registry.FamilyExtra] != 4 || counts[registry.FamilyHighscore] != 1 {
		t.Fatalf("unexpected manifest totals: total=%d families=%v", len(manifest), counts)
	}
	registered, err := registry.Canonical()
	if err != nil {
		t.Fatalf("Canonical() error = %v", err)
	}
	descriptor, found := registered.Resolve("wf_act_toggle_state_random")
	if !found || descriptor.Key != "wf_act_toggle_to_rnd" {
		t.Fatalf("alias resolved to %+v, %t", descriptor, found)
	}
	extension, found := registered.Resolve("wf_cnd_valid_moves")
	if !found || extension.Editor || extension.Family != registry.FamilyCondition || len(registry.CompatibilityManifest()) != 5 {
		t.Fatalf("compatibility extension resolved to %+v, %t", extension, found)
	}
}

// TestRegistryRejectsDuplicatesAndFreezes verifies construction guardrails.
func TestRegistryRejectsDuplicatesAndFreezes(t *testing.T) {
	registered := registry.New()
	descriptor := registry.Descriptor{Key: "wf_test", Family: registry.FamilyTrigger}
	if err := registered.Register(descriptor); err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if err := registered.Register(descriptor); !errors.Is(err, registry.ErrDuplicateKey) {
		t.Fatalf("duplicate error = %v", err)
	}
	registered.Freeze()
	if err := registered.Register(registry.Descriptor{Key: "wf_late", Family: registry.FamilyEffect}); !errors.Is(err, registry.ErrFrozen) {
		t.Fatalf("frozen error = %v", err)
	}
}
