package config

import "testing"

// TestLoadUsesDefaultsAndParsesRarityPolicy verifies safe environment behavior.
func TestLoadUsesDefaultsAndParsesRarityPolicy(t *testing.T) {
	t.Setenv("PIXELS_CRAFTING_RECYCLER_BATCH_SIZE", "bad")
	t.Setenv("PIXELS_CRAFTING_RECYCLER_RARITY_CHANCES", "5=10,2=3")
	config := Load()
	if config.RecyclerBatchSize != 8 || config.RecyclerRarityChance[5] != 10 || config.RecyclerRarityChance[2] != 3 {
		t.Fatalf("config %+v", config)
	}
}

// TestLoadRejectsMalformedRarityPolicy verifies whole-map fallback.
func TestLoadRejectsMalformedRarityPolicy(t *testing.T) {
	t.Setenv("PIXELS_CRAFTING_RECYCLER_RARITY_CHANCES", "5=0")
	config := Load()
	if config.RecyclerRarityChance[5] != 1000 {
		t.Fatalf("config %+v", config)
	}
}
