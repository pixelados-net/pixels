package breeding

import (
	"sort"

	petrecord "github.com/niflaot/pixels/internal/realm/pet/record"
	petreference "github.com/niflaot/pixels/internal/realm/pet/reference"
)

// breedCandidate stores one appearance and its compiled reference weight.
type breedCandidate struct {
	// breed stores the enabled appearance.
	breed petrecord.Breed
	// weight stores its selection weight.
	weight uint64
}

// selectOffspringBreed chooses one enabled breed reproducibly for an operation seed.
func selectOffspringBreed(references *petreference.Snapshot, typeID int32, seed uint64) (petrecord.Breed, bool) {
	candidates := make([]breedCandidate, 0, 16)
	if typeID >= 0 && typeID < int32(len(references.BreedingRaces)) && len(references.BreedingRaces[typeID]) > 0 {
		for _, race := range references.BreedingRaces[typeID] {
			breed, found := references.Breeds[petreference.BreedKey{TypeID: typeID, BreedID: race.BreedID, PaletteID: race.PaletteID}]
			if found && breed.Enabled && race.Enabled && race.Weight > 0 {
				candidates = append(candidates, breedCandidate{breed: breed, weight: uint64(race.Weight)})
			}
		}
	} else {
		for _, breed := range references.Breeds {
			if breed.TypeID == typeID && breed.Enabled {
				candidates = append(candidates, breedCandidate{breed: breed, weight: rarityWeight(breed.Rarity)})
			}
		}
	}
	sort.Slice(candidates, func(first int, second int) bool {
		if candidates[first].breed.Rarity != candidates[second].breed.Rarity {
			return candidates[first].breed.Rarity < candidates[second].breed.Rarity
		}
		if candidates[first].breed.BreedID != candidates[second].breed.BreedID {
			return candidates[first].breed.BreedID < candidates[second].breed.BreedID
		}
		return candidates[first].breed.PaletteID < candidates[second].breed.PaletteID
	})
	total := uint64(0)
	for _, candidate := range candidates {
		total += candidate.weight
	}
	if total == 0 {
		return petrecord.Breed{}, false
	}
	ticket := mixSeed(seed) % total
	for _, candidate := range candidates {
		if ticket < candidate.weight {
			return candidate.breed, true
		}
		ticket -= candidate.weight
	}
	return petrecord.Breed{}, false
}

// rarityWeight returns a stable inverse-rarity selection weight.
func rarityWeight(rarity int32) uint64 {
	if rarity < 0 {
		rarity = 0
	}
	return uint64(max(1, 100/(rarity+1)))
}

// mixSeed avalanches one persisted operation identity into a stable ticket.
func mixSeed(value uint64) uint64 {
	value ^= value >> 30
	value *= 0xbf58476d1ce4e5b9
	value ^= value >> 27
	value *= 0x94d049bb133111eb
	return value ^ value>>31
}
