// Package roller coordinates autonomous room roller furniture.
package roller

import (
	"time"

	"github.com/caarlos0/env/v11"
)

// Config stores roller timing and placement policy.
type Config struct {
	// Delay stores the gap before walk transition hooks run.
	Delay time.Duration `env:"PIXELS_ROLLER_HOOK_DELAY" envDefault:"400ms"`
	// MaxAvatarsPerTick limits player units moved by each roller per cycle.
	MaxAvatarsPerTick int `env:"PIXELS_ROLLER_MAX_AVATARS" envDefault:"1"`
	// NoRules disables roller placement and chain restrictions.
	NoRules bool `env:"PIXELS_ROLLER_NO_RULES" envDefault:"false"`
}

// LoadConfig loads roller policy from environment variables.
func LoadConfig() (Config, error) {
	return env.ParseAs[Config]()
}

// Normalize returns safe roller policy defaults.
func (config Config) Normalize() Config {
	if config.Delay <= 0 {
		config.Delay = 400 * time.Millisecond
	}
	if config.MaxAvatarsPerTick <= 0 {
		config.MaxAvatarsPerTick = 1
	}
	return config
}
