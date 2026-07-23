package config

import (
	"testing"
	"time"
)

// TestLoad verifies progression defaults and environment overrides.
func TestLoad(t *testing.T) {
	t.Setenv("PIXELS_PROGRESSION_ENABLED", "false")
	t.Setenv("PIXELS_PROGRESSION_GUIDE_MIN_TRACK_LEVEL", "3")
	t.Setenv("PIXELS_PROGRESSION_PRESENCE_INTERVAL", "1m")
	loaded := Load()
	if loaded.Enabled || loaded.GuideMinimumTrackLevel != 3 || loaded.PresenceInterval != time.Minute || loaded.FlushInterval != 2*time.Second {
		t.Fatalf("config %+v", loaded)
	}
}
