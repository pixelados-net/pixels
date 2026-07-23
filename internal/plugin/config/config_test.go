package config

import (
	"testing"
	"time"
)

// TestLoadUsesPluginEnvironment verifies plugin runtime overrides.
func TestLoadUsesPluginEnvironment(t *testing.T) {
	t.Setenv("PIXELS_PLUGIN_DIRECTORY", "extensions")
	t.Setenv("PIXELS_PLUGIN_CALLBACK_TIMEOUT", "750ms")
	t.Setenv("PIXELS_COMMAND_PREFIX", "!")
	config, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if config.Directory != "extensions" || config.CallbackTimeout != 750*time.Millisecond || config.CommandPrefix != "!" {
		t.Fatalf("unexpected plugin config %#v", config)
	}
}

// TestNormalizeUsesSafeDefaults verifies invalid programmatic settings.
func TestNormalizeUsesSafeDefaults(t *testing.T) {
	config := (Config{}).Normalize()
	if config.Directory != "plugins" || config.CallbackTimeout != 2*time.Second || config.CommandPrefix != ":" {
		t.Fatalf("unexpected defaults %#v", config)
	}
}
