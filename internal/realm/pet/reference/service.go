// Package reference owns immutable pet species, breed, command, and product snapshots.
package reference

import (
	"context"
	"sync"

	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
)

// Snapshot stores one immutable reference generation.
type Snapshot struct {
	// Species stores species by type identifier.
	Species [36]petrecord.Species
	// SpeciesPresent reports loaded slots without map lookups.
	SpeciesPresent [36]bool
	// Breeds stores appearance records by composite key.
	Breeds map[BreedKey]petrecord.Breed
	// Commands stores commands by protocol identifier.
	Commands [47]petrecord.Command
	// CommandPresent reports loaded commands without map lookups.
	CommandPresent [47]bool
	// SpeciesCommands stores ordered command IDs by species.
	SpeciesCommands [36][]int32
	// ProductRules stores typed products by furniture definition.
	ProductRules map[int64]petrecord.ProductRule
	// Vocals stores ordered localized speech by species.
	Vocals [36][]petrecord.Vocal
	// BreedingRules stores canonical parent compatibility rules.
	BreedingRules map[BreedingKey]petrecord.BreedingRule
	// BreedingRaces stores weighted result appearances by species.
	BreedingRaces [36][]petrecord.BreedingRace
}

// BreedKey identifies one appearance reference record.
type BreedKey struct {
	// TypeID identifies the species.
	TypeID int32
	// BreedID identifies the breed.
	BreedID int32
	// PaletteID identifies the palette.
	PaletteID int32
}

// BreedingKey identifies one canonical parent species pair.
type BreedingKey struct {
	// ParentOneTypeID identifies the lower parent type.
	ParentOneTypeID int32
	// ParentTwoTypeID identifies the higher parent type.
	ParentTwoTypeID int32
}

// Service loads and publishes immutable reference generations.
type Service struct {
	// store supplies reference rows.
	store petrecord.Store
	// mutex protects the current snapshot pointer.
	mutex sync.RWMutex
	// snapshot stores the current immutable generation.
	snapshot *Snapshot
}

// Reader resolves the current immutable reference generation.
type Reader interface {
	// Current returns the current validated snapshot.
	Current(context.Context) (*Snapshot, error)
	// Refresh reloads and publishes one validated snapshot.
	Refresh(context.Context) error
}

// New creates a reference service.
func New(store petrecord.Store) *Service { return &Service{store: store} }

// Refresh loads and atomically publishes one complete reference generation.
func (service *Service) Refresh(ctx context.Context) error {
	species, err := service.store.Species(ctx)
	if err != nil {
		return err
	}
	breeds, err := service.store.Breeds(ctx)
	if err != nil {
		return err
	}
	commands, bySpecies, err := service.store.Commands(ctx)
	if err != nil {
		return err
	}
	products, err := service.store.ProductRules(ctx)
	if err != nil {
		return err
	}
	vocals, err := service.store.Vocals(ctx)
	if err != nil {
		return err
	}
	breedingRules, breedingRaces, err := service.store.BreedingRules(ctx)
	if err != nil {
		return err
	}
	if err = validate(species, breeds, commands, bySpecies, products, vocals, breedingRules, breedingRaces); err != nil {
		return err
	}
	next := &Snapshot{Breeds: make(map[BreedKey]petrecord.Breed, len(breeds)), ProductRules: make(map[int64]petrecord.ProductRule, len(products)), BreedingRules: make(map[BreedingKey]petrecord.BreedingRule, len(breedingRules))}
	for _, item := range species {
		if item.TypeID >= 0 && item.TypeID < int32(len(next.Species)) {
			next.Species[item.TypeID], next.SpeciesPresent[item.TypeID] = item, true
		}
	}
	for _, item := range breeds {
		next.Breeds[BreedKey{TypeID: item.TypeID, BreedID: item.BreedID, PaletteID: item.PaletteID}] = item
	}
	for _, item := range commands {
		if item.ID >= 0 && item.ID < int32(len(next.Commands)) {
			next.Commands[item.ID], next.CommandPresent[item.ID] = item, true
		}
	}
	for typeID, values := range bySpecies {
		if typeID >= 0 && typeID < int32(len(next.SpeciesCommands)) {
			next.SpeciesCommands[typeID] = values
		}
	}
	for _, item := range products {
		next.ProductRules[item.DefinitionID] = item
	}
	for _, item := range vocals {
		next.Vocals[item.TypeID] = append(next.Vocals[item.TypeID], item)
	}
	for _, item := range breedingRules {
		next.BreedingRules[breedingKey(item.ParentOneTypeID, item.ParentTwoTypeID)] = item
	}
	for _, item := range breedingRaces {
		next.BreedingRaces[item.ResultTypeID] = append(next.BreedingRaces[item.ResultTypeID], item)
	}
	service.mutex.Lock()
	service.snapshot = next
	service.mutex.Unlock()
	return nil
}

// Current returns the current generation and lazily loads it once when absent.
func (service *Service) Current(ctx context.Context) (*Snapshot, error) {
	service.mutex.RLock()
	current := service.snapshot
	service.mutex.RUnlock()
	if current != nil {
		return current, nil
	}
	if err := service.Refresh(ctx); err != nil {
		return nil, err
	}
	service.mutex.RLock()
	current = service.snapshot
	service.mutex.RUnlock()
	return current, nil
}

// BreedingResult returns one enabled compatibility result for two parent species.
func (snapshot *Snapshot) BreedingResult(firstTypeID int32, secondTypeID int32) (int32, bool) {
	rule, found := snapshot.BreedingRules[breedingKey(firstTypeID, secondTypeID)]
	return rule.ResultTypeID, found && rule.Enabled
}

// breedingKey canonicalizes one parent pair.
func breedingKey(firstTypeID int32, secondTypeID int32) BreedingKey {
	if firstTypeID > secondTypeID {
		firstTypeID, secondTypeID = secondTypeID, firstTypeID
	}
	return BreedingKey{ParentOneTypeID: firstTypeID, ParentTwoTypeID: secondTypeID}
}
