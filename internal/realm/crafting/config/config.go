// Package config loads crafting and recycler runtime policy.
package config

import (
	"os"
	"strconv"
	"strings"
)

// Config stores crafting and recycler runtime policy.
type Config struct {
	// Enabled controls altar crafting behavior.
	Enabled bool
	// RecyclerEnabled controls recycler behavior.
	RecyclerEnabled bool
	// RecyclerBatchSize stores the exact required item count.
	RecyclerBatchSize int
	// RecyclerRarityChance stores independent tier denominators.
	RecyclerRarityChance map[int32]int
}

// Load reads crafting environment settings with safe defaults.
func Load() Config {
	return Config{Enabled: envBool("PIXELS_CRAFTING_ENABLED", true), RecyclerEnabled: envBool("PIXELS_CRAFTING_RECYCLER_ENABLED", true), RecyclerBatchSize: envInt("PIXELS_CRAFTING_RECYCLER_BATCH_SIZE", 8), RecyclerRarityChance: envChances("PIXELS_CRAFTING_RECYCLER_RARITY_CHANCES", map[int32]int{5: 1000, 4: 100, 3: 20, 2: 5})}
}
func envBool(name string, fallback bool) bool {
	value, err := strconv.ParseBool(os.Getenv(name))
	if err != nil {
		return fallback
	}
	return value
}
func envInt(name string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(name))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
func envChances(name string, fallback map[int32]int) map[int32]int {
	value := strings.TrimSpace(os.Getenv(name))
	if value == "" {
		return fallback
	}
	result := make(map[int32]int, 4)
	for _, pair := range strings.Split(value, ",") {
		parts := strings.SplitN(pair, "=", 2)
		if len(parts) != 2 {
			return fallback
		}
		tier, errTier := strconv.Atoi(strings.TrimSpace(parts[0]))
		chance, errChance := strconv.Atoi(strings.TrimSpace(parts[1]))
		if errTier != nil || errChance != nil || tier < 2 || tier > 5 || chance <= 0 {
			return fallback
		}
		result[int32(tier)] = chance
	}
	if len(result) == 0 {
		return fallback
	}
	return result
}
