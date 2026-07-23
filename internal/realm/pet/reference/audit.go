package reference

import (
	"fmt"

	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
)

// validate rejects incomplete or internally inconsistent reference generations.
func validate(species []petrecord.Species, breeds []petrecord.Breed, commands []petrecord.Command, bySpecies map[int32][]int32, products []petrecord.ProductRule, vocals []petrecord.Vocal, breedingRules []petrecord.BreedingRule, breedingRaces []petrecord.BreedingRace) error {
	speciesPresent, err := validateSpecies(species)
	if err != nil {
		return err
	}
	commandPresent, err := validateCommands(commands)
	if err != nil {
		return err
	}
	if err = validateBreedsAndMappings(species, breeds, bySpecies, speciesPresent, commandPresent); err != nil {
		return err
	}
	if err = validateProducts(products, speciesPresent); err != nil {
		return err
	}
	return validateBehaviorReferences(species, breeds, vocals, breedingRules, breedingRaces, speciesPresent)
}

// validateSpecies audits every stable renderer slot.
func validateSpecies(species []petrecord.Species) ([36]bool, error) {
	present := [36]bool{}
	if len(species) != len(present) {
		return present, fmt.Errorf("pet reference audit: expected 36 species slots, got %d", len(species))
	}
	reservedEnabled := false
	for _, item := range species {
		if item.TypeID < 0 || item.TypeID >= int32(len(present)) || present[item.TypeID] || item.Slug == "" || item.MaximumLevel < 1 || item.MaximumLevel > 20 {
			return present, fmt.Errorf("pet reference audit: invalid or duplicate species %d", item.TypeID)
		}
		present[item.TypeID] = true
		if item.TypeID == 13 {
			reservedEnabled = item.Enabled
		}
	}
	if !present[13] || reservedEnabled {
		return present, fmt.Errorf("pet reference audit: historical species slot 13 must remain disabled")
	}
	return present, nil
}

// validateCommands audits every canonical protocol command slot.
func validateCommands(commands []petrecord.Command) ([47]bool, error) {
	present := [47]bool{}
	for _, item := range commands {
		if item.ID < 0 || item.ID >= int32(len(present)) || item.ID == 39 || present[item.ID] || item.NameKey == "" {
			return present, fmt.Errorf("pet reference audit: invalid or duplicate command %d", item.ID)
		}
		present[item.ID] = true
	}
	for id := int32(0); id <= 46; id++ {
		if id != 39 && !present[id] {
			return present, fmt.Errorf("pet reference audit: command %d is missing", id)
		}
	}
	return present, nil
}

// validateBreedsAndMappings audits enabled appearance and command relationships.
func validateBreedsAndMappings(species []petrecord.Species, breeds []petrecord.Breed, bySpecies map[int32][]int32, speciesPresent [36]bool, commandPresent [47]bool) error {
	breedCount := [36]int{}
	breedKeys := make(map[BreedKey]struct{}, len(breeds))
	for _, item := range breeds {
		key := BreedKey{TypeID: item.TypeID, BreedID: item.BreedID, PaletteID: item.PaletteID}
		if item.TypeID < 0 || item.TypeID >= 36 || !speciesPresent[item.TypeID] {
			return fmt.Errorf("pet reference audit: breed references species %d", item.TypeID)
		}
		if _, duplicate := breedKeys[key]; duplicate {
			return fmt.Errorf("pet reference audit: duplicate breed %+v", key)
		}
		breedKeys[key] = struct{}{}
		if item.Enabled {
			breedCount[item.TypeID]++
		}
	}
	for _, item := range species {
		if item.Enabled && breedCount[item.TypeID] == 0 {
			return fmt.Errorf("pet reference audit: enabled species %d has no enabled breed", item.TypeID)
		}
		seen := [47]bool{}
		for _, commandID := range bySpecies[item.TypeID] {
			if commandID < 0 || commandID >= 47 || commandID == 39 || !commandPresent[commandID] || seen[commandID] {
				return fmt.Errorf("pet reference audit: invalid command %d for species %d", commandID, item.TypeID)
			}
			seen[commandID] = true
		}
		if item.Enabled && len(bySpecies[item.TypeID]) == 0 {
			return fmt.Errorf("pet reference audit: enabled species %d has no commands", item.TypeID)
		}
	}
	return nil
}

// validateProducts audits typed furniture handlers and species restrictions.
func validateProducts(products []petrecord.ProductRule, speciesPresent [36]bool) error {
	definitions := make(map[int64]struct{}, len(products))
	for _, item := range products {
		if item.DefinitionID <= 0 || item.TypeID < -1 || item.TypeID >= 36 || item.TypeID >= 0 && !speciesPresent[item.TypeID] || !validProductKind(item.Kind) {
			return fmt.Errorf("pet reference audit: invalid product rule for definition %d", item.DefinitionID)
		}
		if _, duplicate := definitions[item.DefinitionID]; duplicate {
			return fmt.Errorf("pet reference audit: duplicate product definition %d", item.DefinitionID)
		}
		definitions[item.DefinitionID] = struct{}{}
	}
	return nil
}

// validProductKind reports whether one rule has an implemented handler.
func validProductKind(kind string) bool {
	switch kind {
	case "food", "drink", "toy", "nest", "saddle", "revive", "rebreed", "speed", "seed", "package":
		return true
	default:
		return false
	}
}
