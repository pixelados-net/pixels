// Package action coordinates room avatar expressions and idle projection.
package action

import (
	"time"

	"github.com/caarlos0/env/v11"
)

// Config controls automatic room idle projection.
type Config struct {
	// IdleTimeout stores inactivity before AFK projection.
	IdleTimeout time.Duration `env:"PIXELS_ROOM_IDLE_TIMEOUT" envDefault:"5m"`
	// SweepInterval stores idle reconciliation cadence.
	SweepInterval time.Duration `env:"PIXELS_ROOM_IDLE_SWEEP_INTERVAL" envDefault:"1s"`
	// TransitionDelay stores the pause between cancelling and starting actions.
	TransitionDelay time.Duration `env:"PIXELS_ROOM_ACTION_TRANSITION_DELAY" envDefault:"100ms"`
}

// LoadConfig reads room action configuration from environment variables.
func LoadConfig() (Config, error) { return env.ParseAs[Config]() }

// Normalize fills invalid action configuration values.
func (config Config) Normalize() Config {
	if config.IdleTimeout <= 0 {
		config.IdleTimeout = 5 * time.Minute
	}
	if config.SweepInterval <= 0 {
		config.SweepInterval = time.Second
	}
	if config.TransitionDelay <= 0 {
		config.TransitionDelay = 100 * time.Millisecond
	}
	return config
}
