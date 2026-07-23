package identity

import petrecord "github.com/niflaot/pixels/internal/realm/pet/record"

const (
	monsterPlantPartCount    = 12
	monsterPlantPaletteCount = 11
)

// MonsterPlantAppearance derives one complete deterministic renderer genotype.
func MonsterPlantAppearance(seed uint64) []petrecord.AppearancePart {
	parts := make([]petrecord.AppearancePart, 5)
	parts[0] = petrecord.AppearancePart{LayerID: 0, PartID: -1, PaletteID: 10}
	for layer := int32(1); layer < 5; layer++ {
		value := mixMonsterPlant(seed + uint64(layer)*0x9e3779b97f4a7c15)
		parts[layer] = petrecord.AppearancePart{LayerID: layer, PartID: int32(value%monsterPlantPartCount) + 1, PaletteID: int32((value >> 8) % monsterPlantPaletteCount)}
	}
	return parts
}

// MonsterPlantOffspringAppearance inherits each visible layer from one parent.
func MonsterPlantOffspringAppearance(first []petrecord.AppearancePart, second []petrecord.AppearancePart, seed uint64) []petrecord.AppearancePart {
	fallback := MonsterPlantAppearance(seed)
	parts := make([]petrecord.AppearancePart, len(fallback))
	copy(parts, fallback)
	for layer := int32(1); layer < 5; layer++ {
		preferSecond := mixMonsterPlant(seed+uint64(layer))&1 == 1
		primary, secondary := first, second
		if preferSecond {
			primary, secondary = second, first
		}
		if inherited, found := monsterPlantPart(primary, layer); found {
			parts[layer] = inherited
		} else if inherited, found = monsterPlantPart(secondary, layer); found {
			parts[layer] = inherited
		}
	}
	return parts
}

// monsterPlantPart returns one valid inheritable visible component.
func monsterPlantPart(parts []petrecord.AppearancePart, layer int32) (petrecord.AppearancePart, bool) {
	for _, part := range parts {
		if part.LayerID == layer && part.PartID >= 1 && part.PartID <= monsterPlantPartCount && part.PaletteID >= 0 && part.PaletteID < monsterPlantPaletteCount {
			return part, true
		}
	}
	return petrecord.AppearancePart{}, false
}

// mixMonsterPlant diffuses stable operation identifiers across renderer ranges.
func mixMonsterPlant(value uint64) uint64 {
	value ^= value >> 30
	value *= 0xbf58476d1ce4e5b9
	value ^= value >> 27
	value *= 0x94d049bb133111eb
	return value ^ value>>31
}
