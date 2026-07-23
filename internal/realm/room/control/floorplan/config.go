// Package floorplan validates and authorizes custom room floor plans.
package floorplan

import (
	"time"

	"github.com/caarlos0/env/v11"
)

const (
	// MaxDimension stores the largest editable floor plan width or height.
	MaxDimension = 64
	// MaxTiles stores the largest editable rectangular floor plan area.
	MaxTiles = MaxDimension * MaxDimension
	// MaxAutoPickupItems limits furniture returned by one internal save.
	MaxAutoPickupItems = 500
)

// Config controls floor plan validation and save throttling.
type Config struct {
	// RejectZeroEffectiveHeight rejects maps containing no usable tile.
	RejectZeroEffectiveHeight bool `env:"PIXELS_ROOM_FLOORPLAN_REJECT_ZERO_HEIGHT" envDefault:"true"`
	// SaveCooldown stores the minimum interval between player saves.
	SaveCooldown time.Duration `env:"PIXELS_ROOM_FLOORPLAN_SAVE_COOLDOWN" envDefault:"3s"`
}

// LoadConfig reads floor plan configuration from environment variables.
func LoadConfig() (Config, error) {
	return env.ParseAs[Config]()
}

// Normalize fills invalid values with conservative defaults.
func (config Config) Normalize() Config {
	if config.SaveCooldown <= 0 {
		config.SaveCooldown = 3 * time.Second
	}

	return config
}
