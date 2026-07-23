package firework

import (
	"time"

	"github.com/caarlos0/env/v11"
)

const defaultRecharge = 5 * time.Second

// Config stores firework recharge policy.
type Config struct {
	// DefaultRecharge stores the fallback delay when furniture has no valid override.
	DefaultRecharge time.Duration `env:"PIXELS_FIREWORK_DEFAULT_RECHARGE" envDefault:"5s"`
}

// LoadConfig returns environment-backed firework policy.
func LoadConfig() (Config, error) { return env.ParseAs[Config]() }

// Normalize returns safe bounded firework policy.
func (config Config) Normalize() Config {
	if config.DefaultRecharge <= 0 || config.DefaultRecharge > 5*time.Minute {
		config.DefaultRecharge = defaultRecharge
	}

	return config
}
