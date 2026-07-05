package redis

import (
	"os"
	"testing"
)

// TestLoadConfigUsesDefault verifies default Redis configuration.
func TestLoadConfigUsesDefault(t *testing.T) {
	clearEnv(t, "REDIS_ADDRESS", "REDIS_USERNAME", "REDIS_PASSWORD", "REDIS_DATABASE")

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if config.Address != "127.0.0.1:6379" {
		t.Fatalf("expected default address, got %q", config.Address)
	}
}

// TestLoadConfigUsesEnvironment verifies Redis configuration from environment variables.
func TestLoadConfigUsesEnvironment(t *testing.T) {
	t.Setenv("REDIS_ADDRESS", "localhost:6380")
	t.Setenv("REDIS_USERNAME", "default")
	t.Setenv("REDIS_PASSWORD", "secret")
	t.Setenv("REDIS_DATABASE", "2")

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if config.Address != "localhost:6380" {
		t.Fatalf("expected environment address, got %q", config.Address)
	}

	if config.Database != 2 {
		t.Fatalf("expected database 2, got %d", config.Database)
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
