package config

import "testing"

// TestLoad verifies the enabled override and default.
func TestLoad(t *testing.T) {
	t.Setenv("PIXELS_GAMECENTER_ENABLED", "false")
	if Load().Enabled {
		t.Fatal("expected disabled")
	}
	t.Setenv("PIXELS_GAMECENTER_ENABLED", "invalid")
	if !Load().Enabled {
		t.Fatal("expected safe default")
	}
}
