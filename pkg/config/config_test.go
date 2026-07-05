package config

import (
	"os"
	"path/filepath"
	"testing"

	appconfig "github.com/niflaot/pixels/pkg/config/app"
	"github.com/niflaot/pixels/pkg/logger"
	"github.com/niflaot/pixels/pkg/redis"
	"go.uber.org/fx"
	"go.uber.org/fx/fxtest"
)

// TestLoadUsesEnvironment verifies composed configuration from environment variables.
func TestLoadUsesEnvironment(t *testing.T) {
	t.Setenv("PIXELS_ENV", "test")
	t.Setenv("PIXELS_HOST", "0.0.0.0")
	t.Setenv("PIXELS_PORT", "8080")
	t.Setenv("PIXELS_ACCESS_KEY", "secret")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("LOG_FORMAT", "json")
	t.Setenv("REDIS_ADDRESS", "localhost:6380")

	config, err := Load()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if config.App.Environment != "test" {
		t.Fatalf("expected environment test, got %q", config.App.Environment)
	}

	if config.App.Address() != "0.0.0.0:8080" {
		t.Fatalf("expected address 0.0.0.0:8080, got %q", config.App.Address())
	}

	if config.App.AccessKey != "secret" {
		t.Fatalf("expected access key from environment, got %q", config.App.AccessKey)
	}

	if config.Logger.Format != logger.FormatJSON {
		t.Fatalf("expected json logger format, got %q", config.Logger.Format)
	}

	if config.Redis.Address != "localhost:6380" {
		t.Fatalf("expected Redis address from environment, got %q", config.Redis.Address)
	}
}

// TestLoadUsesDotenv verifies dotenv files populate environment variables.
func TestLoadUsesDotenv(t *testing.T) {
	clearEnv(t, "PIXELS_ENV", "PIXELS_HOST", "PIXELS_PORT", "PIXELS_ACCESS_KEY", "LOG_LEVEL", "LOG_FORMAT", "REDIS_ADDRESS")

	path := filepath.Join(t.TempDir(), ".env")
	content := "PIXELS_ENV=dotenv\nPIXELS_HOST=localhost\nPIXELS_PORT=9090\nPIXELS_ACCESS_KEY=dotenv-key\nLOG_LEVEL=warn\nLOG_FORMAT=console\nREDIS_ADDRESS=localhost:6381\n"

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write dotenv: %v", err)
	}

	config, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if config.App.Environment != "dotenv" {
		t.Fatalf("expected dotenv environment, got %q", config.App.Environment)
	}

	if config.App.Port != 9090 {
		t.Fatalf("expected dotenv port 9090, got %d", config.App.Port)
	}

	if config.App.AccessKey != "dotenv-key" {
		t.Fatalf("expected dotenv access key, got %q", config.App.AccessKey)
	}

	if config.Redis.Address != "localhost:6381" {
		t.Fatalf("expected dotenv Redis address, got %q", config.Redis.Address)
	}
}

// TestLoadReturnsDotenvError verifies explicit dotenv load errors are returned.
func TestLoadReturnsDotenvError(t *testing.T) {
	_, err := Load(filepath.Join(t.TempDir(), "missing.env"))
	if err == nil {
		t.Fatal("expected missing dotenv error")
	}
}

// TestLoadReturnsEnvironmentError verifies invalid environment values are returned.
func TestLoadReturnsEnvironmentError(t *testing.T) {
	clearEnv(t, "PIXELS_ENV", "PIXELS_HOST", "PIXELS_ACCESS_KEY", "LOG_LEVEL", "LOG_FORMAT", "REDIS_ADDRESS")
	t.Setenv("PIXELS_PORT", "invalid")

	_, err := Load()
	if err == nil {
		t.Fatal("expected invalid environment error")
	}
}

// TestModuleProvidesConfig verifies the Fx module exposes composed and focused config.
func TestModuleProvidesConfig(t *testing.T) {
	clearEnv(t, "PIXELS_ENV", "PIXELS_HOST", "PIXELS_PORT", "PIXELS_ACCESS_KEY", "LOG_LEVEL", "LOG_FORMAT", "REDIS_ADDRESS")

	var invoked bool
	app := fxtest.New(
		t,
		Module,
		fx.Invoke(func(config AppConfig, app appconfig.Config, log logger.Config, redis redis.Config) {
			invoked = true

			if config.App != app {
				t.Fatalf("expected app config provider to match composed config")
			}

			if config.Logger != log {
				t.Fatalf("expected logger config provider to match composed config")
			}

			if config.Redis != redis {
				t.Fatalf("expected Redis config provider to match composed config")
			}
		}),
	)

	app.RequireStart()
	app.RequireStop()

	if !invoked {
		t.Fatal("expected module invocation")
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
