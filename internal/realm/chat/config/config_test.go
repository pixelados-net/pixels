package config

import (
	"testing"
	"time"
)

// TestConfigNormalize verifies defaults and protocol tier selection.
func TestConfigNormalize(t *testing.T) {
	config := (Config{}).Normalize()
	if config.MaxMessageRunes != 256 || config.HistoryBatchSize != 200 || config.Tier(2).MaxMessages != 6 {
		t.Fatalf("unexpected normalized config: %#v", config)
	}
	if tier := (Config{Tier1MaxMessages: 3, Tier1Window: time.Second}).Tier(1); tier.MaxMessages != 3 || tier.Window != time.Second {
		t.Fatalf("unexpected tier: %#v", tier)
	}
	if AudienceDistance(0) != 50 || AudienceDistance(12) != 12 {
		t.Fatal("unexpected audience distance")
	}
}

// TestLoadConfig verifies representative environment reproduction.
func TestLoadConfig(t *testing.T) {
	t.Setenv("PIXELS_CHAT_MAX_MESSAGE_RUNES", "80")
	t.Setenv("PIXELS_CHAT_FLOOD_TIER2_WINDOW", "9s")
	config, err := LoadConfig()
	if err != nil || config.MaxMessageRunes != 80 || config.Tier2Window != 9*time.Second {
		t.Fatalf("config=%#v err=%v", config, err)
	}
}
