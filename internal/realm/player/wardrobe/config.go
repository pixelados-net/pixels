package wardrobe

import "github.com/caarlos0/env/v11"

// Config contains wardrobe slot policy.
type Config struct {
	// MinimumSlot is the first writable Nitro wardrobe slot.
	MinimumSlot int32 `env:"PIXELS_PLAYER_WARDROBE_MIN_SLOT" envDefault:"1"`
	// MaximumSlot is the last writable Nitro wardrobe slot.
	MaximumSlot int32 `env:"PIXELS_PLAYER_WARDROBE_MAX_SLOT" envDefault:"10"`
}

// LoadConfig loads wardrobe policy from environment variables and defaults.
func LoadConfig() (Config, error) { return env.ParseAs[Config]() }

// DefaultConfig returns the documented wardrobe policy defaults.
func DefaultConfig() Config { return Config{MinimumSlot: MinSlot, MaximumSlot: MaxSlot} }
