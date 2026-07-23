// Package wired owns configuration shared by the room WIRED capability.
package wired

import "github.com/caarlos0/env/v11"

// Config controls WIRED compilation and execution budgets.
type Config struct {
	// Enabled permits WIRED configuration and execution.
	Enabled bool `env:"PIXELS_WIRED_ENABLED" envDefault:"true"`
	// MaxSelection bounds furniture targets stored by one node.
	MaxSelection int `env:"PIXELS_WIRED_MAX_SELECTION" envDefault:"20"`
	// MaxDelayPulses bounds a delayed action in 500 millisecond pulses.
	MaxDelayPulses int32 `env:"PIXELS_WIRED_MAX_DELAY_PULSES" envDefault:"7200"`
	// MaxEventsPerTrace bounds derived events in one execution trace.
	MaxEventsPerTrace int `env:"PIXELS_WIRED_MAX_EVENTS_PER_TRACE" envDefault:"128"`
	// MaxStacksPerTrace bounds stacks visited in one execution trace.
	MaxStacksPerTrace int `env:"PIXELS_WIRED_MAX_STACKS_PER_TRACE" envDefault:"64"`
	// MaxEffectsPerTrace bounds effects executed in one execution trace.
	MaxEffectsPerTrace int `env:"PIXELS_WIRED_MAX_EFFECTS_PER_TRACE" envDefault:"128"`
	// MaxCallDepth bounds nested call-stack requests.
	MaxCallDepth int `env:"PIXELS_WIRED_MAX_CALL_DEPTH" envDefault:"10"`
	// MaxDelayedPerRoom bounds outstanding delayed actions.
	MaxDelayedPerRoom int `env:"PIXELS_WIRED_MAX_DELAYED_PER_ROOM" envDefault:"512"`
	// HighscoreTop bounds retained highscore entries per board and period.
	HighscoreTop int `env:"PIXELS_WIRED_HIGHSCORE_TOP" envDefault:"50"`
}

// LoadConfig reads WIRED configuration from environment variables.
func LoadConfig() (Config, error) { return env.ParseAs[Config]() }

// Normalize replaces invalid budget values with safe defaults.
func (config Config) Normalize() Config {
	if config.MaxSelection <= 0 {
		config.MaxSelection = 20
	}
	if config.MaxDelayPulses <= 0 {
		config.MaxDelayPulses = 7200
	}
	if config.MaxEventsPerTrace <= 0 {
		config.MaxEventsPerTrace = 128
	}
	if config.MaxStacksPerTrace <= 0 {
		config.MaxStacksPerTrace = 64
	}
	if config.MaxEffectsPerTrace <= 0 {
		config.MaxEffectsPerTrace = 128
	}
	if config.MaxCallDepth <= 0 {
		config.MaxCallDepth = 10
	}
	if config.MaxDelayedPerRoom <= 0 {
		config.MaxDelayedPerRoom = 512
	}
	if config.HighscoreTop <= 0 {
		config.HighscoreTop = 50
	}
	return config
}
