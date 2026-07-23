package app

import (
	"os"
	"testing"
)

// TestLoadUsesDefault verifies default application configuration.
func TestLoadUsesDefault(t *testing.T) {
	clearEnv(t, "PIXELS_ENV", "PIXELS_HOST", "PIXELS_PORT", "PIXELS_ACCESS_KEY")

	config, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if config.Environment != "development" {
		t.Fatalf("expected development environment, got %q", config.Environment)
	}

	if config.Address() != "127.0.0.1:3000" {
		t.Fatalf("expected default address, got %q", config.Address())
	}

	if config.AccessKey != "pixels-development-access-key-change-me" {
		t.Fatalf("expected default access key, got %q", config.AccessKey)
	}
}

// TestLoadUsesEnvironment verifies application configuration from the environment.
func TestLoadUsesEnvironment(t *testing.T) {
	t.Setenv("PIXELS_ENV", "test")
	t.Setenv("PIXELS_HOST", "0.0.0.0")
	t.Setenv("PIXELS_PORT", "8080")
	t.Setenv("PIXELS_ACCESS_KEY", "secret")

	config, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if config.Environment != "test" {
		t.Fatalf("expected test environment, got %q", config.Environment)
	}

	if config.Address() != "0.0.0.0:8080" {
		t.Fatalf("expected environment address, got %q", config.Address())
	}

	if config.AccessKey != "secret" {
		t.Fatalf("expected environment access key, got %q", config.AccessKey)
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
