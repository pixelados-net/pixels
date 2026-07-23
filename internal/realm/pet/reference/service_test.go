package reference

import (
	"context"
	"strings"
	"testing"
	"time"

	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
)

// referenceStore supplies one complete in-memory reference generation.
type referenceStore struct {
	petrecord.Store
	// species stores test species.
	species []petrecord.Species
	// breeds stores test breeds.
	breeds []petrecord.Breed
	// commands stores test commands.
	commands []petrecord.Command
	// speciesCommands stores test associations.
	speciesCommands map[int32][]int32
	// products stores test product rules.
	products []petrecord.ProductRule
	// vocals stores test localized speech.
	vocals []petrecord.Vocal
	// breedingRules stores test parent compatibility.
	breedingRules []petrecord.BreedingRule
	// breedingRaces stores test offspring appearances.
	breedingRaces []petrecord.BreedingRace
}

// Species returns test species.
func (store referenceStore) Species(context.Context) ([]petrecord.Species, error) {
	return store.species, nil
}

// Breeds returns test breeds.
func (store referenceStore) Breeds(context.Context) ([]petrecord.Breed, error) {
	return store.breeds, nil
}

// Commands returns test commands and associations.
func (store referenceStore) Commands(context.Context) ([]petrecord.Command, map[int32][]int32, error) {
	return store.commands, store.speciesCommands, nil
}

// ProductRules returns test product rules.
func (store referenceStore) ProductRules(context.Context) ([]petrecord.ProductRule, error) {
	return store.products, nil
}

// Vocals returns test localized speech.
func (store referenceStore) Vocals(context.Context) ([]petrecord.Vocal, error) {
	return store.vocals, nil
}

// BreedingRules returns test compatibility and appearance rules.
func (store referenceStore) BreedingRules(context.Context) ([]petrecord.BreedingRule, []petrecord.BreedingRace, error) {
	return store.breedingRules, store.breedingRaces, nil
}

// TestBreedKeyComparable verifies the composite key is map-safe and stable.
func TestBreedKeyComparable(t *testing.T) {
	values := map[BreedKey]bool{{TypeID: 1, BreedID: 2, PaletteID: 3}: true}
	if !values[BreedKey{TypeID: 1, BreedID: 2, PaletteID: 3}] {
		t.Fatal("expected matching composite key")
	}
}

// TestRefreshPublishesCompleteGeneration verifies the startup reference gate.
func TestRefreshPublishesCompleteGeneration(t *testing.T) {
	store := completeReferenceStore()
	service := New(store)
	if err := service.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	current, err := service.Current(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !current.SpeciesPresent[35] || !current.CommandPresent[46] || current.CommandPresent[39] {
		t.Fatalf("unexpected published manifest: species=%v command46=%v command39=%v", current.SpeciesPresent[35], current.CommandPresent[46], current.CommandPresent[39])
	}
}

// TestRefreshRejectsIncompleteReferences verifies corrupt generations never replace current data.
func TestRefreshRejectsIncompleteReferences(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*referenceStore)
		want   string
	}{
		{name: "species slot", mutate: func(store *referenceStore) { store.species = store.species[:35] }, want: "36 species"},
		{name: "reserved enabled", mutate: func(store *referenceStore) { store.species[13].Enabled = true }, want: "slot 13"},
		{name: "command missing", mutate: func(store *referenceStore) { store.commands = store.commands[:45] }, want: "command 46"},
		{name: "breed missing", mutate: func(store *referenceStore) { store.breeds = store.breeds[:35] }, want: "no enabled breed"},
		{name: "mapping missing", mutate: func(store *referenceStore) { store.speciesCommands[0] = nil }, want: "no commands"},
		{name: "product kind", mutate: func(store *referenceStore) { store.products[0].Kind = "unknown" }, want: "product rule"},
		{name: "vocal missing", mutate: func(store *referenceStore) { store.vocals = store.vocals[1:] }, want: "no vocal"},
		{name: "breeding missing", mutate: func(store *referenceStore) { store.breedingRules = store.breedingRules[1:] }, want: "breeding rule"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			store := completeReferenceStore()
			test.mutate(&store)
			err := New(store).Refresh(context.Background())
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("expected %q audit error, got %v", test.want, err)
			}
		})
	}
}

// completeReferenceStore creates every canonical species and command slot.
func completeReferenceStore() referenceStore {
	store := referenceStore{speciesCommands: make(map[int32][]int32, 36), products: []petrecord.ProductRule{{DefinitionID: 1, Kind: "food", TypeID: -1, Enabled: true}}}
	commandIDs := make([]int32, 0, 46)
	for id := int32(0); id <= 46; id++ {
		if id == 39 {
			continue
		}
		store.commands = append(store.commands, petrecord.Command{ID: id, NameKey: "command"})
		commandIDs = append(commandIDs, id)
	}
	for typeID := int32(0); typeID < 36; typeID++ {
		enabled := typeID != 13
		store.species = append(store.species, petrecord.Species{TypeID: typeID, Slug: "species", MaximumLevel: 20, Breedable: enabled, Enabled: enabled})
		store.breeds = append(store.breeds, petrecord.Breed{TypeID: typeID, Enabled: enabled})
		if enabled {
			store.speciesCommands[typeID] = append([]int32(nil), commandIDs...)
			store.vocals = append(store.vocals, petrecord.Vocal{TypeID: typeID, Mood: "idle", TextKey: "pet.vocal.generic", Weight: 1, Cooldown: 15 * time.Second, Enabled: true})
			store.breedingRules = append(store.breedingRules, petrecord.BreedingRule{ParentOneTypeID: typeID, ParentTwoTypeID: typeID, ResultTypeID: typeID, Enabled: true})
			store.breedingRaces = append(store.breedingRaces, petrecord.BreedingRace{ResultTypeID: typeID, Weight: 100, Enabled: true})
		}
	}
	return store
}
