package wired

import "testing"

// TestLoadConfigReadsDefaultsAndOverrides verifies every runtime budget has a usable environment contract.
func TestLoadConfigReadsDefaultsAndOverrides(t *testing.T) {
	config, err := LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if !config.Enabled || config.MaxSelection != 20 || config.MaxDelayPulses != 7200 || config.HighscoreTop != 50 {
		t.Fatalf("default config=%+v", config)
	}
	t.Setenv("PIXELS_WIRED_ENABLED", "false")
	t.Setenv("PIXELS_WIRED_MAX_SELECTION", "7")
	t.Setenv("PIXELS_WIRED_MAX_DELAY_PULSES", "8")
	t.Setenv("PIXELS_WIRED_MAX_EVENTS_PER_TRACE", "9")
	t.Setenv("PIXELS_WIRED_MAX_STACKS_PER_TRACE", "10")
	t.Setenv("PIXELS_WIRED_MAX_EFFECTS_PER_TRACE", "11")
	t.Setenv("PIXELS_WIRED_MAX_CALL_DEPTH", "12")
	t.Setenv("PIXELS_WIRED_MAX_DELAYED_PER_ROOM", "13")
	t.Setenv("PIXELS_WIRED_HIGHSCORE_TOP", "14")
	config, err = LoadConfig()
	if err != nil {
		t.Fatal(err)
	}
	if config.Enabled || config.MaxSelection != 7 || config.MaxDelayPulses != 8 || config.MaxEventsPerTrace != 9 || config.MaxStacksPerTrace != 10 || config.MaxEffectsPerTrace != 11 || config.MaxCallDepth != 12 || config.MaxDelayedPerRoom != 13 || config.HighscoreTop != 14 {
		t.Fatalf("override config=%+v", config)
	}
}

// TestNormalizeReplacesEveryInvalidBudget verifies zero values cannot disable safety limits.
func TestNormalizeReplacesEveryInvalidBudget(t *testing.T) {
	config := (Config{Enabled: true}).Normalize()
	if config.MaxSelection != 20 || config.MaxDelayPulses != 7200 || config.MaxEventsPerTrace != 128 || config.MaxStacksPerTrace != 64 || config.MaxEffectsPerTrace != 128 || config.MaxCallDepth != 10 || config.MaxDelayedPerRoom != 512 || config.HighscoreTop != 50 {
		t.Fatalf("normalized config=%+v", config)
	}
}
