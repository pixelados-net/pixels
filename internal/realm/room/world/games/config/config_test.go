package config

import "testing"

// TestLoad verifies documented overrides and safe defaults.
func TestLoad(t *testing.T) {
	t.Setenv("PIXELS_GAMES_ENABLED", "false")
	t.Setenv("PIXELS_GAMES_FREEZE_MAX_LIVES", "4")
	t.Setenv("PIXELS_GAMES_BANZAI_POINTS_LOCK", "2")
	loaded := Load()
	if loaded.Enabled || loaded.Freeze.MaxLives != 4 || loaded.Banzai.PointsLock != 2 {
		t.Fatalf("unexpected config: %+v", loaded)
	}
}
