package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/niflaot/pixels/internal/auth/sso"
	pluginconfig "github.com/niflaot/pixels/internal/plugin/config"
	currencyconfig "github.com/niflaot/pixels/internal/realm/inventory/currency"
	roomentry "github.com/niflaot/pixels/internal/realm/room/access/entry"
	roommoderation "github.com/niflaot/pixels/internal/realm/room/control/moderation"
	appconfig "github.com/niflaot/pixels/pkg/config/app"
	"github.com/niflaot/pixels/pkg/i18n"
	"github.com/niflaot/pixels/pkg/logger"
	"github.com/niflaot/pixels/pkg/postgres"
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
	t.Setenv("TOON_CONSOLE", "true")
	t.Setenv("PIXELS_I18N_PATH", "custom-i18n.json")
	t.Setenv("PIXELS_CURRENCY_TYPES", "-1:credits,5:diamonds")
	t.Setenv("PIXELS_CURRENCY_LEDGER_TYPES", "-1,5")
	t.Setenv("PIXELS_PLUGIN_DIRECTORY", "extensions")
	t.Setenv("PIXELS_PLUGIN_CALLBACK_TIMEOUT", "750ms")
	t.Setenv("PIXELS_COMMAND_PREFIX", "!")
	t.Setenv("PIXELS_POSTGRES_HOST", "db")
	t.Setenv("PIXELS_POSTGRES_DATABASE", "pixels_test")
	t.Setenv("REDIS_ADDRESS", "localhost:6380")
	t.Setenv("SSO_DEFAULT_TTL", "10m")
	t.Setenv("SSO_KEY", "secret-sso-key")
	t.Setenv("PIXELS_ROOM_MODERATION_MIN_MUTE_MINUTES", "2")
	t.Setenv("PIXELS_ROOM_MODERATION_MAX_MUTE_MINUTES", "120")

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

	if !config.Logger.ToonConsole {
		t.Fatal("expected toon console from environment")
	}

	if config.I18N.Path != "custom-i18n.json" {
		t.Fatalf("expected i18n path from environment, got %q", config.I18N.Path)
	}
	if config.Currency.Types != "-1:credits,5:diamonds" || len(config.Currency.LedgerTypes) != 2 {
		t.Fatalf("unexpected currency config %#v", config.Currency)
	}
	if config.Plugin.Directory != "extensions" || config.Plugin.CommandPrefix != "!" {
		t.Fatalf("unexpected plugin config %#v", config.Plugin)
	}

	if config.Postgres.Database != "pixels_test" {
		t.Fatalf("expected PostgreSQL database from environment, got %q", config.Postgres.Database)
	}

	if config.Redis.Address != "localhost:6380" {
		t.Fatalf("expected Redis address from environment, got %q", config.Redis.Address)
	}

	if config.SSO.Key != "secret-sso-key" {
		t.Fatalf("expected SSO key from environment, got %q", config.SSO.Key)
	}
	if config.RoomModeration.MinMuteMinutes != 2 || config.RoomModeration.MaxMuteMinutes != 120 {
		t.Fatalf("unexpected room moderation config %#v", config.RoomModeration)
	}
}

// TestLoadUsesDotenv verifies dotenv files populate environment variables.
func TestLoadUsesDotenv(t *testing.T) {
	clearEnv(t, "PIXELS_ENV", "PIXELS_HOST", "PIXELS_PORT", "PIXELS_ACCESS_KEY", "LOG_LEVEL", "LOG_FORMAT", "TOON_CONSOLE", "PIXELS_I18N_PATH", "PIXELS_CURRENCY_TYPES", "PIXELS_CURRENCY_LEDGER_TYPES", "PIXELS_PLUGIN_DIRECTORY", "PIXELS_PLUGIN_CALLBACK_TIMEOUT", "PIXELS_COMMAND_PREFIX", "PIXELS_POSTGRES_HOST", "REDIS_ADDRESS", "SSO_DEFAULT_TTL", "SSO_KEY")

	path := filepath.Join(t.TempDir(), ".env")
	content := "PIXELS_ENV=dotenv\nPIXELS_HOST=localhost\nPIXELS_PORT=9090\nPIXELS_ACCESS_KEY=dotenv-key\nLOG_LEVEL=warn\nLOG_FORMAT=console\nTOON_CONSOLE=true\nPIXELS_I18N_PATH=dotenv-i18n.json\nPIXELS_CURRENCY_TYPES=-1:credits,5:diamonds\nPIXELS_CURRENCY_LEDGER_TYPES=-1,5\nPIXELS_PLUGIN_DIRECTORY=dotenv-plugins\nPIXELS_PLUGIN_CALLBACK_TIMEOUT=900ms\nPIXELS_COMMAND_PREFIX=!\nPIXELS_POSTGRES_HOST=dotenv-db\nREDIS_ADDRESS=localhost:6381\nSSO_DEFAULT_TTL=15m\nSSO_KEY=dotenv-sso-key\n"

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

	if !config.Logger.ToonConsole {
		t.Fatal("expected dotenv toon console")
	}

	if config.I18N.Path != "dotenv-i18n.json" {
		t.Fatalf("expected dotenv i18n path, got %q", config.I18N.Path)
	}
	if config.Currency.Types != "-1:credits,5:diamonds" || len(config.Currency.LedgerTypes) != 2 {
		t.Fatalf("unexpected dotenv currency config %#v", config.Currency)
	}
	if config.Plugin.Directory != "dotenv-plugins" || config.Plugin.CommandPrefix != "!" {
		t.Fatalf("unexpected dotenv plugin config %#v", config.Plugin)
	}

	if config.Postgres.Host != "dotenv-db" {
		t.Fatalf("expected dotenv PostgreSQL host, got %q", config.Postgres.Host)
	}

	if config.Redis.Address != "localhost:6381" {
		t.Fatalf("expected dotenv Redis address, got %q", config.Redis.Address)
	}

	if config.SSO.Key != "dotenv-sso-key" {
		t.Fatalf("expected dotenv SSO key, got %q", config.SSO.Key)
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
	clearEnv(t, "PIXELS_ENV", "PIXELS_HOST", "PIXELS_ACCESS_KEY", "LOG_LEVEL", "LOG_FORMAT", "TOON_CONSOLE", "PIXELS_I18N_PATH", "REDIS_ADDRESS", "SSO_DEFAULT_TTL")
	t.Setenv("PIXELS_PORT", "invalid")

	_, err := Load()
	if err == nil {
		t.Fatal("expected invalid environment error")
	}
}

// TestModuleProvidesConfig verifies the Fx module exposes composed and focused config.
func TestModuleProvidesConfig(t *testing.T) {
	clearEnv(t, "PIXELS_ENV", "PIXELS_HOST", "PIXELS_PORT", "PIXELS_ACCESS_KEY", "LOG_LEVEL", "LOG_FORMAT", "TOON_CONSOLE", "PIXELS_I18N_PATH", "PIXELS_CURRENCY_TYPES", "PIXELS_CURRENCY_LEDGER_TYPES", "PIXELS_PLUGIN_DIRECTORY", "PIXELS_PLUGIN_CALLBACK_TIMEOUT", "PIXELS_COMMAND_PREFIX", "PIXELS_POSTGRES_HOST", "REDIS_ADDRESS", "SSO_DEFAULT_TTL", "SSO_KEY")

	var invoked bool
	app := fxtest.New(
		t,
		Module,
		fx.Invoke(func(config AppConfig, app appconfig.Config, log logger.Config, translations i18n.Config, currency currencyconfig.Config, plugin pluginconfig.Config, entry roomentry.Config, moderation roommoderation.Config, postgres postgres.Config, redis redis.Config, sso sso.Config) {
			invoked = true

			if config.App != app {
				t.Fatalf("expected app config provider to match composed config")
			}

			if config.Logger != log {
				t.Fatalf("expected logger config provider to match composed config")
			}

			if config.I18N != translations {
				t.Fatalf("expected i18n config provider to match composed config")
			}
			if config.Currency.Types != currency.Types {
				t.Fatalf("expected currency config provider to match composed config")
			}
			if config.Plugin != plugin {
				t.Fatalf("expected plugin config provider to match composed config")
			}
			if config.RoomEntry != entry {
				t.Fatalf("expected room entry config provider to match composed config")
			}
			if config.RoomModeration != moderation {
				t.Fatalf("expected room moderation config provider to match composed config")
			}

			if config.Postgres != postgres {
				t.Fatalf("expected PostgreSQL config provider to match composed config")
			}

			if config.Redis != redis {
				t.Fatalf("expected Redis config provider to match composed config")
			}

			if config.SSO != sso {
				t.Fatalf("expected SSO config provider to match composed config")
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
