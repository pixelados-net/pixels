package reference

import (
	"fmt"
	"time"

	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
)

// validateBehaviorReferences audits vocal and breeding relationship manifests.
func validateBehaviorReferences(species []petrecord.Species, breeds []petrecord.Breed, vocals []petrecord.Vocal, rules []petrecord.BreedingRule, races []petrecord.BreedingRace, speciesPresent [36]bool) error {
	vocalCount := [36]int{}
	for _, item := range vocals {
		if item.TypeID < 0 || item.TypeID >= 36 || !speciesPresent[item.TypeID] || item.Mood == "" || item.TextKey == "" || item.Weight < 1 || item.Cooldown < time.Second || item.Cooldown > time.Hour {
			return fmt.Errorf("pet reference audit: invalid vocal for species %d", item.TypeID)
		}
		if item.Enabled {
			vocalCount[item.TypeID]++
		}
	}
	breedKeys := make(map[BreedKey]bool, len(breeds))
	for _, item := range breeds {
		breedKeys[BreedKey{TypeID: item.TypeID, BreedID: item.BreedID, PaletteID: item.PaletteID}] = item.Enabled
	}
	ruleKeys := make(map[BreedingKey]struct{}, len(rules))
	resultCount := [36]int{}
	for _, item := range rules {
		key := breedingKey(item.ParentOneTypeID, item.ParentTwoTypeID)
		if key.ParentOneTypeID != item.ParentOneTypeID || key.ParentTwoTypeID != item.ParentTwoTypeID || item.ParentOneTypeID < 0 || item.ParentOneTypeID >= 36 || item.ParentTwoTypeID < 0 || item.ParentTwoTypeID >= 36 || item.ResultTypeID < 0 || item.ResultTypeID >= 36 || !speciesPresent[item.ParentOneTypeID] || !speciesPresent[item.ParentTwoTypeID] || !speciesPresent[item.ResultTypeID] {
			return fmt.Errorf("pet reference audit: invalid breeding rule %d/%d", item.ParentOneTypeID, item.ParentTwoTypeID)
		}
		if _, duplicate := ruleKeys[key]; duplicate {
			return fmt.Errorf("pet reference audit: duplicate breeding rule %d/%d", item.ParentOneTypeID, item.ParentTwoTypeID)
		}
		ruleKeys[key] = struct{}{}
	}
	raceKeys := make(map[BreedKey]struct{}, len(races))
	for _, item := range races {
		key := BreedKey{TypeID: item.ResultTypeID, BreedID: item.BreedID, PaletteID: item.PaletteID}
		if item.ResultTypeID < 0 || item.ResultTypeID >= 36 || !breedKeys[key] || item.Weight < 1 {
			return fmt.Errorf("pet reference audit: invalid breeding race %+v", key)
		}
		if _, duplicate := raceKeys[key]; duplicate {
			return fmt.Errorf("pet reference audit: duplicate breeding race %+v", key)
		}
		raceKeys[key] = struct{}{}
		if item.Enabled {
			resultCount[item.ResultTypeID]++
		}
	}
	for _, item := range species {
		if item.Enabled && vocalCount[item.TypeID] == 0 {
			return fmt.Errorf("pet reference audit: enabled species %d has no vocal", item.TypeID)
		}
		if !item.Enabled || !item.Breedable {
			continue
		}
		rule, found := findSelfBreedingRule(rules, item.TypeID)
		if !found || !rule.Enabled || resultCount[rule.ResultTypeID] == 0 {
			return fmt.Errorf("pet reference audit: breedable species %d has no complete breeding rule", item.TypeID)
		}
	}
	return nil
}

// findSelfBreedingRule returns one same-species compatibility row.
func findSelfBreedingRule(rules []petrecord.BreedingRule, typeID int32) (petrecord.BreedingRule, bool) {
	for _, item := range rules {
		if item.ParentOneTypeID == typeID && item.ParentTwoTypeID == typeID {
			return item, true
		}
	}
	return petrecord.BreedingRule{}, false
}

// TrainingCommands returns all compatible commands and the currently learned subset.
func TrainingCommands(snapshot *Snapshot, pet petrecord.Pet) ([]int32, []int32) {
	var all []int32
	if pet.TypeID >= 0 && pet.TypeID < int32(len(snapshot.SpeciesCommands)) {
		all = snapshot.SpeciesCommands[pet.TypeID]
	}
	enabled := make([]int32, 0, len(all))
	for _, commandID := range all {
		if commandID >= 0 && commandID < int32(len(snapshot.Commands)) && snapshot.CommandPresent[commandID] && snapshot.Commands[commandID].Enabled && snapshot.Commands[commandID].RequiredLevel <= pet.Level {
			enabled = append(enabled, commandID)
		}
	}
	return all, enabled
}
