package record

import "time"

// LevelThresholds stores cumulative experience required for levels two through twenty.
var LevelThresholds = [...]int32{100, 200, 400, 600, 900, 1300, 1800, 2400, 3200, 4300, 5700, 7600, 10100, 13300, 17500, 23000, 30200, 39600, 51900}

// LevelForExperience returns a monotonic level bounded by maximumLevel.
func LevelForExperience(experience int32, maximumLevel int32) int32 {
	if maximumLevel <= 0 {
		maximumLevel = 20
	}
	level := int32(1)
	for _, threshold := range LevelThresholds {
		if experience < threshold || level >= maximumLevel {
			break
		}
		level++
	}
	return level
}

// NextThreshold returns the next experience goal or the final threshold at cap.
func NextThreshold(level int32) int32 {
	if level <= 0 {
		return LevelThresholds[0]
	}
	index := int(level - 1)
	if index >= len(LevelThresholds) {
		return LevelThresholds[len(LevelThresholds)-1]
	}
	return LevelThresholds[index]
}

// MaximumEnergy returns the level-scaled energy cap.
func MaximumEnergy(level int32) int32 {
	if level < 1 {
		level = 1
	}
	return level * 100
}

// MaterializeStats derives current needs from one value and absolute timestamp.
func MaterializeStats(pet Pet, now time.Time, interval time.Duration, energyDecay int32, happinessDecay int32) Pet {
	if interval <= 0 || pet.StatsAt.IsZero() || !now.After(pet.StatsAt) {
		return pet
	}
	steps := int32(now.Sub(pet.StatsAt) / interval)
	if steps <= 0 {
		return pet
	}
	pet.Energy = clampStat(pet.Energy-steps*energyDecay, MaximumEnergy(pet.Level))
	pet.Happiness = clampStat(pet.Happiness-steps*happinessDecay, 100)
	pet.StatsAt = pet.StatsAt.Add(time.Duration(steps) * interval)
	return pet
}

// clampStat bounds one materialized need to its current cap.
func clampStat(value int32, maximum int32) int32 {
	if value < 0 {
		return 0
	}
	if value > maximum {
		return maximum
	}
	return value
}
