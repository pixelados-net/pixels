package bundle

import "github.com/caarlos0/env/v11"

// Config stores room bundle cloning policy.
type Config struct {
	// CloneBots reports whether template bots are copied into purchased rooms.
	CloneBots bool `env:"PIXELS_ROOM_BUNDLE_BOTS_ENABLED" envDefault:"true"`
}

// LoadConfig reads room bundle cloning policy.
func LoadConfig() (Config, error) { return env.ParseAs[Config]() }
