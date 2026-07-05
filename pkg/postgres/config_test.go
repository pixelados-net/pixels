package postgres

import (
	"os"
	"strings"
	"testing"
	"time"
)

// TestLoadConfigUsesDefault verifies default PostgreSQL configuration.
func TestLoadConfigUsesDefault(t *testing.T) {
	clearEnv(t, postgresEnvKeys()...)

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if config.Host != "localhost" {
		t.Fatalf("expected default host, got %q", config.Host)
	}

	if config.Port != 5432 {
		t.Fatalf("expected default port, got %d", config.Port)
	}

	if config.StatementTimeout != 5*time.Second {
		t.Fatalf("expected default statement timeout, got %s", config.StatementTimeout)
	}
}

// TestLoadConfigUsesEnvironment verifies PostgreSQL configuration from environment variables.
func TestLoadConfigUsesEnvironment(t *testing.T) {
	t.Setenv("PIXELS_POSTGRES_HOST", "db")
	t.Setenv("PIXELS_POSTGRES_PORT", "15432")
	t.Setenv("PIXELS_POSTGRES_DATABASE", "pixels_test")
	t.Setenv("PIXELS_POSTGRES_USER", "tester")
	t.Setenv("PIXELS_POSTGRES_PASSWORD", "secret")
	t.Setenv("PIXELS_POSTGRES_SSL_MODE", "require")
	t.Setenv("PIXELS_POSTGRES_MAX_CONNS", "20")
	t.Setenv("PIXELS_POSTGRES_MIN_CONNS", "2")
	t.Setenv("PIXELS_POSTGRES_CONNECT_TIMEOUT", "3s")
	t.Setenv("PIXELS_POSTGRES_STATEMENT_TIMEOUT", "4s")
	t.Setenv("PIXELS_POSTGRES_HEALTH_TIMEOUT", "1s")

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if config.Database != "pixels_test" {
		t.Fatalf("expected database from environment, got %q", config.Database)
	}

	if config.MaxConns != 20 {
		t.Fatalf("expected max conns from environment, got %d", config.MaxConns)
	}

	if config.HealthTimeout != time.Second {
		t.Fatalf("expected health timeout from environment, got %s", config.HealthTimeout)
	}
}

// TestDSNMasksPassword verifies masked DSNs do not expose secrets.
func TestDSNMasksPassword(t *testing.T) {
	config := Config{
		Host:           "localhost",
		Port:           5432,
		Database:       "pixels",
		User:           "pixels",
		Password:       "super-secret",
		SSLMode:        "disable",
		ConnectTimeout: 5 * time.Second,
	}

	if !strings.Contains(config.DSN(), "super-secret") {
		t.Fatal("expected DSN to contain configured password")
	}

	if strings.Contains(config.MaskedDSN(), "super-secret") {
		t.Fatal("expected masked DSN to hide configured password")
	}
}

// postgresEnvKeys returns PostgreSQL environment variable names.
func postgresEnvKeys() []string {
	return []string{
		"PIXELS_POSTGRES_HOST",
		"PIXELS_POSTGRES_PORT",
		"PIXELS_POSTGRES_DATABASE",
		"PIXELS_POSTGRES_USER",
		"PIXELS_POSTGRES_PASSWORD",
		"PIXELS_POSTGRES_SSL_MODE",
		"PIXELS_POSTGRES_MAX_CONNS",
		"PIXELS_POSTGRES_MIN_CONNS",
		"PIXELS_POSTGRES_CONNECT_TIMEOUT",
		"PIXELS_POSTGRES_STATEMENT_TIMEOUT",
		"PIXELS_POSTGRES_HEALTH_TIMEOUT",
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
