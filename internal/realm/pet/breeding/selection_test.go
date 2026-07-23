package breeding

import (
	"testing"

	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	petreference "github.com/niflaot/pixels/internal/realm/pet/reference"
)

// TestSelectOffspringBreedDeterministic verifies map order cannot change a retry result.
func TestSelectOffspringBreedDeterministic(t *testing.T) {
	first := snapshotWithBreeds([]petrecord.Breed{{TypeID: 1, BreedID: 2, PaletteID: 1, Rarity: 2, Enabled: true}, {TypeID: 1, BreedID: 1, PaletteID: 1, Enabled: true}})
	second := snapshotWithBreeds([]petrecord.Breed{{TypeID: 1, BreedID: 1, PaletteID: 1, Enabled: true}, {TypeID: 1, BreedID: 2, PaletteID: 1, Rarity: 2, Enabled: true}})
	want, found := selectOffspringBreed(first, 1, 42)
	got, secondFound := selectOffspringBreed(second, 1, 42)
	if !found || !secondFound || got != want {
		t.Fatalf("selection mismatch: got %+v, want %+v", got, want)
	}
}

// TestSelectOffspringBreedUsesConfiguredWeights verifies compiled race weights override rarity defaults.
func TestSelectOffspringBreedUsesConfiguredWeights(t *testing.T) {
	references := snapshotWithBreeds([]petrecord.Breed{{TypeID: 1, BreedID: 1, PaletteID: 1, Enabled: true}, {TypeID: 1, BreedID: 2, PaletteID: 1, Enabled: true}})
	references.BreedingRaces[1] = []petrecord.BreedingRace{{ResultTypeID: 1, BreedID: 1, PaletteID: 1, Weight: 1, Enabled: true}, {ResultTypeID: 1, BreedID: 2, PaletteID: 1, Weight: 1000, Enabled: true}}
	selected, found := selectOffspringBreed(references, 1, 1)
	if !found || selected.BreedID != 2 {
		t.Fatalf("expected weighted breed 2, got %+v found=%v", selected, found)
	}
}

// BenchmarkBreedSelect measures the non-hot breeding decision allocation budget.
func BenchmarkBreedSelect(b *testing.B) {
	references := snapshotWithBreeds([]petrecord.Breed{{TypeID: 1, BreedID: 1, PaletteID: 1, Enabled: true}, {TypeID: 1, BreedID: 2, PaletteID: 2, Rarity: 3, Enabled: true}})
	b.ReportAllocs()
	for index := 0; index < b.N; index++ {
		_, _ = selectOffspringBreed(references, 1, uint64(index))
	}
}

// snapshotWithBreeds builds one reference snapshot for selection tests.
func snapshotWithBreeds(values []petrecord.Breed) *petreference.Snapshot {
	result := &petreference.Snapshot{Breeds: make(map[petreference.BreedKey]petrecord.Breed, len(values))}
	for _, value := range values {
		result.Breeds[petreference.BreedKey{TypeID: value.TypeID, BreedID: value.BreedID, PaletteID: value.PaletteID}] = value
	}
	return result
}
