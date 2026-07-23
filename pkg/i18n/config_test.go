package i18n

import "testing"

// TestLoadConfigUsesDefaults verifies default i18n configuration.
func TestLoadConfigUsesDefaults(t *testing.T) {
	clearEnv(t, "PIXELS_I18N_PATH", "PIXELS_I18N_DEFAULT_LOCALE", "PIXELS_I18N_FALLBACK_LOCALE", "PIXELS_I18N_MISSING_MODE")

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if config.Path != "i18n/translations.json" || config.DefaultLocale != "es" || config.FallbackLocale != "en" || config.MissingMode != MissingKey {
		t.Fatalf("unexpected config %#v", config)
	}
}

// TestLoadConfigUsesEnvironment verifies environment overrides.
func TestLoadConfigUsesEnvironment(t *testing.T) {
	t.Setenv("PIXELS_I18N_PATH", "custom.json")
	t.Setenv("PIXELS_I18N_DEFAULT_LOCALE", "en")
	t.Setenv("PIXELS_I18N_FALLBACK_LOCALE", "es")
	t.Setenv("PIXELS_I18N_MISSING_MODE", "empty")

	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if config.Path != "custom.json" || config.DefaultLocale != "en" || config.FallbackLocale != "es" || config.MissingMode != MissingEmpty {
		t.Fatalf("unexpected config %#v", config)
	}
}
