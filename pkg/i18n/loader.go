package i18n

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"go.uber.org/zap"
)

// document stores the JSON translation file shape.
type document struct {
	// Version stores the catalog file version.
	Version int `json:"version"`
	// Locales stores translations by locale and key.
	Locales map[string]map[string]string `json:"locales"`
}

// LoadCatalog reads the configured translation catalog.
func LoadCatalog(config Config, log *zap.Logger) (*Catalog, error) {
	config = config.Normalize()
	data, err := os.ReadFile(config.Path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if log != nil {
				log.Warn("i18n catalog missing", zap.String("path", config.Path))
			}

			return NewCatalog(config, nil), nil
		}

		return nil, fmt.Errorf("read i18n catalog: %w", err)
	}

	entries, err := parseCatalog(data)
	if err != nil {
		return nil, fmt.Errorf("parse i18n catalog: %w", err)
	}

	if log != nil {
		log.Info("i18n catalog loaded", zap.String("path", config.Path), zap.Int("locales", len(entries)), zap.Int("keys", countEntries(entries)))
	}

	return NewCatalog(config, entries), nil
}

// NewTranslator exposes a catalog as translator.
func NewTranslator(catalog *Catalog) Translator {
	return catalog
}

// parseCatalog decodes catalog JSON.
func parseCatalog(data []byte) (map[Locale]map[Key]string, error) {
	var raw document
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}

	entries := make(map[Locale]map[Key]string, len(raw.Locales))
	for locale, values := range raw.Locales {
		keyed := make(map[Key]string, len(values))
		for key, value := range values {
			keyed[Key(key)] = value
		}
		entries[Locale(locale)] = keyed
	}

	return entries, nil
}

// countEntries counts translation keys.
func countEntries(entries map[Locale]map[Key]string) int {
	total := 0
	for _, values := range entries {
		total += len(values)
	}

	return total
}
