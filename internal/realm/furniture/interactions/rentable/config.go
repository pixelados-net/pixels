package rentable

import (
	"time"

	"github.com/caarlos0/env/v11"
)

// Config controls rentable furniture prices and duration.
type Config struct {
	// Duration stores one rental extension period.
	Duration time.Duration `env:"PIXELS_RENTABLE_DURATION" envDefault:"24h"`
	// PriceCredits stores one rental extension price.
	PriceCredits int32 `env:"PIXELS_RENTABLE_PRICE_CREDITS" envDefault:"10"`
	// BuyoutCredits stores permanent ownership price.
	BuyoutCredits int32 `env:"PIXELS_RENTABLE_BUYOUT_CREDITS" envDefault:"50"`
}

// LoadConfig reads rentable furniture settings.
func LoadConfig() (Config, error) { return env.ParseAs[Config]() }

// Normalize fills invalid rentable settings.
func (config Config) Normalize() Config {
	if config.Duration <= 0 {
		config.Duration = 24 * time.Hour
	}
	if config.PriceCredits < 0 {
		config.PriceCredits = 10
	}
	if config.BuyoutCredits < 0 {
		config.BuyoutCredits = 50
	}
	return config
}
