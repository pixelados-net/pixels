// Package moderation coordinates persistent room moderation.
package moderation

import "github.com/caarlos0/env/v11"

const (
	// defaultMinMuteMinutes stores the minimum mute duration.
	defaultMinMuteMinutes int32 = 1
	// defaultMaxMuteMinutes stores the maximum mute duration.
	defaultMaxMuteMinutes int32 = 1440
)

// Config controls room moderation limits.
type Config struct {
	// MinMuteMinutes stores the minimum accepted mute duration.
	MinMuteMinutes int32 `env:"PIXELS_ROOM_MODERATION_MIN_MUTE_MINUTES" envDefault:"1"`
	// MaxMuteMinutes stores the maximum accepted mute duration.
	MaxMuteMinutes int32 `env:"PIXELS_ROOM_MODERATION_MAX_MUTE_MINUTES" envDefault:"1440"`
}

// LoadConfig reads room moderation configuration.
func LoadConfig() (Config, error) {
	return env.ParseAs[Config]()
}

// Normalize fills invalid moderation limits with defaults.
func (config Config) Normalize() Config {
	if config.MinMuteMinutes <= 0 {
		config.MinMuteMinutes = defaultMinMuteMinutes
	}
	if config.MaxMuteMinutes < config.MinMuteMinutes {
		config.MaxMuteMinutes = defaultMaxMuteMinutes
	}

	return config
}
