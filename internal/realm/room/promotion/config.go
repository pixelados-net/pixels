package promotion

import (
	"time"

	"github.com/caarlos0/env/v11"
)

// Config controls purchased room promotion duration.
type Config struct {
	// Duration stores the time added by each purchase.
	Duration time.Duration `env:"PIXELS_ROOM_PROMOTION_DURATION" envDefault:"2h"`
}

// LoadConfig loads promotion settings from environment variables.
func LoadConfig() (Config, error) { return env.ParseAs[Config]() }

// Normalize restores invalid promotion settings.
func (config Config) Normalize() Config {
	if config.Duration <= 0 {
		config.Duration = 2 * time.Hour
	}
	return config
}
