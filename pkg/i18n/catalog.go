package i18n

import "strings"

// Locale names a translation locale.
type Locale string

// Key names one translation entry.
type Key string

// Params stores interpolation replacements.
type Params map[string]string

// Translator resolves localized text.
type Translator interface {
	// T resolves a key for a locale.
	T(Locale, Key, ...Params) string
	// Default resolves a key for the configured default locale.
	Default(Key, ...Params) string
	// Entries returns one locale's resolved translation entries.
	Entries(Locale) map[Key]string
}

// Entries returns one locale's resolved translation entries.
func (catalog *Catalog) Entries(locale Locale) map[Key]string {
	if catalog == nil {
		return map[Key]string{}
	}
	if locale == "" {
		locale = catalog.defaultLocale
	}
	values := catalog.entries[locale]
	if values == nil {
		values = catalog.entries[catalog.fallbackLocale]
	}
	entries := make(map[Key]string, len(values))
	for key, value := range values {
		entries[key] = value
	}

	return entries
}

// Catalog stores immutable translation entries.
type Catalog struct {
	// defaultLocale stores the primary locale.
	defaultLocale Locale
	// fallbackLocale stores the secondary locale.
	fallbackLocale Locale
	// missingMode stores missing key behavior.
	missingMode MissingMode
	// entries stores translations by locale and key.
	entries map[Locale]map[Key]string
}

// NewCatalog creates a translation catalog.
func NewCatalog(config Config, entries map[Locale]map[Key]string) *Catalog {
	config = config.Normalize()

	return &Catalog{
		defaultLocale:  config.DefaultLocale,
		fallbackLocale: config.FallbackLocale,
		missingMode:    config.MissingMode,
		entries:        copyEntries(entries),
	}
}

// T resolves a key for a locale.
func (catalog *Catalog) T(locale Locale, key Key, params ...Params) string {
	if catalog == nil {
		return string(key)
	}
	if locale == "" {
		locale = catalog.defaultLocale
	}

	text, found := catalog.find(locale, key)
	if !found && catalog.fallbackLocale != "" && catalog.fallbackLocale != locale {
		text, found = catalog.find(catalog.fallbackLocale, key)
	}
	if !found {
		if catalog.missingMode == MissingEmpty {
			return ""
		}

		text = string(key)
	}

	return interpolate(text, params...)
}

// Default resolves a key for the configured default locale.
func (catalog *Catalog) Default(key Key, params ...Params) string {
	if catalog == nil {
		return string(key)
	}

	return catalog.T(catalog.defaultLocale, key, params...)
}

// find returns one translation entry.
func (catalog *Catalog) find(locale Locale, key Key) (string, bool) {
	entries := catalog.entries[locale]
	if entries == nil {
		return "", false
	}

	value, found := entries[key]

	return value, found
}

// interpolate replaces simple placeholders in text.
func interpolate(text string, params ...Params) string {
	for _, replacements := range params {
		for key, value := range replacements {
			text = strings.ReplaceAll(text, "{"+key+"}", value)
		}
	}

	return text
}

// copyEntries copies catalog entries.
func copyEntries(entries map[Locale]map[Key]string) map[Locale]map[Key]string {
	copied := make(map[Locale]map[Key]string, len(entries))
	for locale, values := range entries {
		copiedValues := make(map[Key]string, len(values))
		for key, value := range values {
			copiedValues[key] = value
		}
		copied[locale] = copiedValues
	}

	return copied
}
