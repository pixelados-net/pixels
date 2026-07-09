// Package currency contains inventory currency configuration and catalog data.
package currency

import "github.com/caarlos0/env/v11"

// Config contains currency catalog and ledger settings.
type Config struct {
	// CatalogPath stores the JSON currency type catalog path.
	CatalogPath string `env:"PIXELS_CURRENCY_CATALOG_PATH" envDefault:"currency/types.json"`

	// LedgerTypes stores currency types that require audit entries.
	LedgerTypes []int32 `env:"PIXELS_CURRENCY_LEDGER_TYPES" envDefault:"-1" envSeparator:","`
}

// LoadConfig reads currency configuration from environment variables.
func LoadConfig() (Config, error) {
	return env.ParseAs[Config]()
}
