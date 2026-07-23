// Package i18n contains reusable translation catalog loading and lookup.
package i18n

import "github.com/caarlos0/env/v11"

const (
	// MissingKey returns the missing translation key.
	MissingKey MissingMode = "key"

	// MissingEmpty returns an empty string for missing translations.
	MissingEmpty MissingMode = "empty"
)

// MissingMode names missing translation behavior.
type MissingMode string

// Config contains i18n catalog settings.
type Config struct {
	// Path stores the JSON translation catalog path.
	Path string `env:"PIXELS_I18N_PATH" envDefault:"i18n/translations.json"`

	// DefaultLocale stores the locale used without player preference.
	DefaultLocale Locale `env:"PIXELS_I18N_DEFAULT_LOCALE" envDefault:"es"`

	// FallbackLocale stores the secondary lookup locale.
	FallbackLocale Locale `env:"PIXELS_I18N_FALLBACK_LOCALE" envDefault:"en"`

	// MissingMode stores behavior for untranslated keys.
	MissingMode MissingMode `env:"PIXELS_I18N_MISSING_MODE" envDefault:"key"`
}

// LoadConfig reads i18n configuration from environment variables.
func LoadConfig() (Config, error) {
	return env.ParseAs[Config]()
}

// Normalize fills invalid optional settings with defaults.
func (config Config) Normalize() Config {
	if config.Path == "" {
		config.Path = "i18n/translations.json"
	}
	if config.DefaultLocale == "" {
		config.DefaultLocale = "es"
	}
	if config.FallbackLocale == "" {
		config.FallbackLocale = "en"
	}
	if config.MissingMode != MissingEmpty {
		config.MissingMode = MissingKey
	}

	return config
}
