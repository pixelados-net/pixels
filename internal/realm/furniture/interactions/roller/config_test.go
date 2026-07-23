package roller

import (
	"testing"
	"time"
)

// TestConfigLoadsAndNormalizes verifies environment parsing and safe defaults.
func TestConfigLoadsAndNormalizes(t *testing.T) {
	t.Setenv("PIXELS_ROLLER_HOOK_DELAY", "250ms")
	t.Setenv("PIXELS_ROLLER_MAX_AVATARS", "2")
	t.Setenv("PIXELS_ROLLER_NO_RULES", "true")
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if config.Delay != 250*time.Millisecond || config.MaxAvatarsPerTick != 2 || !config.NoRules {
		t.Fatalf("unexpected config %#v", config)
	}
	normalized := (Config{}).Normalize()
	if normalized.Delay != 400*time.Millisecond || normalized.MaxAvatarsPerTick != 1 {
		t.Fatalf("unexpected defaults %#v", normalized)
	}
}
