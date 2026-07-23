// Package builders coordinates Builders Club purchase-and-place requests.
package builders

import "github.com/caarlos0/env/v11"

// Config controls the deliberately disabled Builders Club placement policy.
type Config struct {
	// FurnitureLimit stores the maximum placed furniture count; zero disables the tier.
	FurnitureLimit int `env:"PIXELS_BUILDERS_CLUB_FURNITURE_LIMIT" envDefault:"0"`
}

// LoadConfig reads Builders Club policy from the environment.
func LoadConfig() (Config, error) { return env.ParseAs[Config]() }

// Normalize clamps invalid policy to the disabled state.
func (config Config) Normalize() Config {
	if config.FurnitureLimit < 0 {
		config.FurnitureLimit = 0
	}

	return config
}
