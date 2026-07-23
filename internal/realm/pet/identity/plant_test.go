package identity

import (
	"reflect"
	"testing"

	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
)

// TestMonsterPlantAppearanceIsCompleteAndDeterministic verifies valid stable genetics.
func TestMonsterPlantAppearanceIsCompleteAndDeterministic(t *testing.T) {
	first := MonsterPlantAppearance(500008)
	if !reflect.DeepEqual(first, MonsterPlantAppearance(500008)) {
		t.Fatal("expected deterministic genetics")
	}
	if len(first) != 5 || reflect.DeepEqual(first, MonsterPlantAppearance(500009)) {
		t.Fatalf("unexpected genetics %+v", first)
	}
	if first[0] != (petrecord.AppearancePart{LayerID: 0, PartID: -1, PaletteID: 10}) {
		t.Fatalf("unexpected pot %+v", first[0])
	}
	for layer, part := range first[1:] {
		if part.LayerID != int32(layer+1) || part.PartID < 1 || part.PartID > 12 || part.PaletteID < 0 || part.PaletteID > 10 {
			t.Fatalf("invalid part %+v", part)
		}
	}
}

// TestMonsterPlantOffspringAppearanceInheritsVisibleLayers verifies parental genes win over fallbacks.
func TestMonsterPlantOffspringAppearanceInheritsVisibleLayers(t *testing.T) {
	first := MonsterPlantAppearance(1)
	second := MonsterPlantAppearance(2)
	offspring := MonsterPlantOffspringAppearance(first, second, 3)
	for layer := 1; layer < len(offspring); layer++ {
		if offspring[layer] != first[layer] && offspring[layer] != second[layer] {
			t.Fatalf("layer %d was not inherited: %+v", layer, offspring[layer])
		}
	}
}
