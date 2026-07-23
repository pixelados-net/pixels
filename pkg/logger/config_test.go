package logger

import (
	"os"
	"testing"
)

// TestLoadConfigUsesDefault verifies default logger configuration.
func TestLoadConfigUsesDefault(t *testing.T) {
	clearEnv(t, "LOG_LEVEL", "LOG_FORMAT", "TOON_CONSOLE")

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if config.Level != "info" {
		t.Fatalf("expected info level, got %q", config.Level)
	}

	if config.Format != FormatConsole {
		t.Fatalf("expected console format, got %q", config.Format)
	}

	if config.ToonConsole {
		t.Fatal("expected toon console disabled")
	}
}

// TestLoadConfigUsesEnvironment verifies logger configuration from the environment.
func TestLoadConfigUsesEnvironment(t *testing.T) {
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("LOG_FORMAT", "json")
	t.Setenv("TOON_CONSOLE", "true")

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if config.Level != "debug" {
		t.Fatalf("expected debug level, got %q", config.Level)
	}

	if config.Format != FormatJSON {
		t.Fatalf("expected json format, got %q", config.Format)
	}

	if !config.ToonConsole {
		t.Fatal("expected toon console enabled")
	}
}

// clearEnv removes environment variables for a test and restores them afterward.
func clearEnv(t *testing.T, keys ...string) {
	t.Helper()

	values := make(map[string]string, len(keys))
	present := make(map[string]bool, len(keys))

	for _, key := range keys {
		value, ok := os.LookupEnv(key)
		values[key] = value
		present[key] = ok
		_ = os.Unsetenv(key)
	}

	t.Cleanup(func() {
		for _, key := range keys {
			if present[key] {
				_ = os.Setenv(key, values[key])
				continue
			}

			_ = os.Unsetenv(key)
		}
	})
}
