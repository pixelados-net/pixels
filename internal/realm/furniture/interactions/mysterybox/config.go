package mysterybox

import (
	"time"

	"github.com/caarlos0/env/v11"
)

// Config controls the deterministic default mystery-box prize pool.
type Config struct {
	// PrizeDefinitionID identifies the granted default prize.
	PrizeDefinitionID int64 `env:"PIXELS_MYSTERYBOX_PRIZE_DEFINITION_ID" envDefault:"1"`
	// Wait stores the client-visible reveal delay.
	Wait time.Duration `env:"PIXELS_MYSTERYBOX_WAIT" envDefault:"3s"`
}

// LoadConfig reads mystery-box settings.
func LoadConfig() (Config, error) { return env.ParseAs[Config]() }

// Normalize fills invalid mystery-box settings.
func (config Config) Normalize() Config {
	if config.PrizeDefinitionID <= 0 {
		config.PrizeDefinitionID = 1
	}
	if config.Wait <= 0 {
		config.Wait = 3 * time.Second
	}
	return config
}
