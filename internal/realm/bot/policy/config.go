// Package policy owns bot limits and permission capabilities.
package policy

import (
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
)

// Config stores bounded bot behavior settings.
type Config struct {
	// MaxPerRoom stores the ordinary room bot limit.
	MaxPerRoom int `env:"PIXELS_BOT_MAX_PER_ROOM" envDefault:"25"`
	// MaxInventory stores the ordinary per-player inventory limit.
	MaxInventory int `env:"PIXELS_BOT_MAX_INVENTORY" envDefault:"25"`
	// WalkRadius stores maximum random-walk distance from current tile.
	WalkRadius int `env:"PIXELS_BOT_WALK_RADIUS" envDefault:"5"`
	// LimitWalkRadius constrains random walking around the current tile.
	LimitWalkRadius bool `env:"PIXELS_BOT_LIMIT_WALK_RADIUS" envDefault:"true"`
	// BartenderCommandDistance stores maximum keyword hearing distance.
	BartenderCommandDistance int `env:"PIXELS_BOT_BARTENDER_COMMAND_DISTANCE" envDefault:"6"`
	// BartenderReachDistance stores immediate hand-item delivery distance.
	BartenderReachDistance int `env:"PIXELS_BOT_BARTENDER_REACH_DISTANCE" envDefault:"3"`
	// PlacementMessages stores semicolon-separated localized placement keys.
	PlacementMessages string `env:"PIXELS_BOT_PLACEMENT_MESSAGES" envDefault:"bots.placement.hello;bots.placement.party;bots.placement.welcome"`
	// PositionFlushInterval stores deferred position persistence cadence.
	PositionFlushInterval time.Duration `env:"PIXELS_BOT_POSITION_FLUSH_INTERVAL" envDefault:"5s"`
}

// LoadConfig loads bot settings from environment variables.
func LoadConfig() (Config, error) {
	config := Config{}
	if err := env.Parse(&config); err != nil {
		return Config{}, err
	}
	return config.Normalize(), nil
}

// Normalize clamps unsafe bot configuration values.
func (config Config) Normalize() Config {
	if config.MaxPerRoom <= 0 {
		config.MaxPerRoom = 25
	}
	if config.MaxInventory <= 0 {
		config.MaxInventory = 25
	}
	if config.WalkRadius <= 0 {
		config.WalkRadius = 5
	}
	if config.BartenderCommandDistance <= 0 {
		config.BartenderCommandDistance = 6
	}
	if config.BartenderReachDistance <= 0 {
		config.BartenderReachDistance = 3
	}
	if config.PositionFlushInterval <= 0 {
		config.PositionFlushInterval = 5 * time.Second
	}
	return config
}

// PlacementKeys returns non-empty configured translation keys.
func (config Config) PlacementKeys() []string {
	parts := strings.Split(config.PlacementMessages, ";")
	keys := parts[:0]
	for _, part := range parts {
		if key := strings.TrimSpace(part); key != "" {
			keys = append(keys, key)
		}
	}
	return keys
}
