package floorplan

import (
	"testing"
	"time"
)

// TestLoadConfigReadsFloorplanEnvironment verifies floor plan environment reproduction.
func TestLoadConfigReadsFloorplanEnvironment(t *testing.T) {
	t.Setenv("PIXELS_ROOM_FLOORPLAN_REJECT_ZERO_HEIGHT", "false")
	t.Setenv("PIXELS_ROOM_FLOORPLAN_SAVE_COOLDOWN", "7s")
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if config.RejectZeroEffectiveHeight || config.SaveCooldown != 7*time.Second {
		t.Fatalf("unexpected config %#v", config)
	}
}

// TestConfigNormalizeRestoresCooldown verifies conservative invalid duration handling.
func TestConfigNormalizeRestoresCooldown(t *testing.T) {
	if (Config{}).Normalize().SaveCooldown != 3*time.Second {
		t.Fatal("expected default cooldown")
	}
}
