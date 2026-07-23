package behavior

import (
	"testing"

	petruntime "github.com/niflaot/pixels/internal/realm/pet/runtime"
)

// TestRegistryCoversCanonicalCommands verifies all 46 non-gap command IDs.
func TestRegistryCoversCanonicalCommands(t *testing.T) {
	registry := NewRegistry()
	if err := registry.Validate(); err != nil {
		t.Fatal(err)
	}
	if _, found := registry.Find(39); found {
		t.Fatal("historical command gap 39 must remain unregistered")
	}
	if value, found := registry.Resolve("sígueme"); !found || value.ID != 7 {
		t.Fatalf("unexpected localized alias %#v found=%t", value, found)
	}
}

// TestRegistryUsesContextualFoodCommands verifies eat and drink require matching products.
func TestRegistryUsesContextualFoodCommands(t *testing.T) {
	registry := NewRegistry()
	tests := []struct {
		name string
		id   int32
		need petruntime.CommandNeed
	}{
		{name: "drink", id: 14, need: petruntime.CommandNeedDrink},
		{name: "eat", id: 43, need: petruntime.CommandNeedFood},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			definition, found := registry.Find(test.id)
			if !found || definition.Action.Mode != petruntime.ActionNeed || definition.Action.Need != test.need {
				t.Fatalf("definition=%+v found=%v", definition, found)
			}
		})
	}
}
