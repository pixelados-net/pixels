package builders

import "testing"

// TestConfigNormalizePreservesDisabledPolicy verifies zero remains deliberate.
func TestConfigNormalizePreservesDisabledPolicy(t *testing.T) {
	if value := (Config{}).Normalize().FurnitureLimit; value != 0 {
		t.Fatalf("limit=%d", value)
	}
	if value := (Config{FurnitureLimit: -1}).Normalize().FurnitureLimit; value != 0 {
		t.Fatalf("negative limit=%d", value)
	}
	if value := (Config{FurnitureLimit: 25}).Normalize().FurnitureLimit; value != 25 {
		t.Fatalf("enabled limit=%d", value)
	}
}

// BenchmarkConfigNormalize measures the dormant policy hot check.
func BenchmarkConfigNormalize(b *testing.B) {
	config := Config{FurnitureLimit: 25}
	for range b.N {
		_ = config.Normalize()
	}
}
