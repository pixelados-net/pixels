package settings

import (
	"time"

	"github.com/caarlos0/env/v11"
)

// Config contains client-settings write coalescing policy.
type Config struct {
	// FlushInterval controls final settings persistence cadence.
	FlushInterval time.Duration `env:"PIXELS_PLAYER_SETTINGS_FLUSH_INTERVAL" envDefault:"250ms"`
	// PendingLimit bounds distinct players waiting for persistence.
	PendingLimit int `env:"PIXELS_PLAYER_SETTINGS_PENDING_LIMIT" envDefault:"4096"`
}

// LoadConfig loads client-settings policy from environment variables and defaults.
func LoadConfig() (Config, error) { return env.ParseAs[Config]() }
