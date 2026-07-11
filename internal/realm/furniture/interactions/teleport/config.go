// Package teleport implements paired furniture teleport behavior.
package teleport

import "github.com/caarlos0/env/v11"

// Config stores furniture teleport policy.
type Config struct {
	// BypassLocked allows cross-room teleports through locked room modes.
	BypassLocked bool `env:"PIXELS_FURNITURE_TELEPORT_BYPASS_LOCKED" envDefault:"false"`
}

// LoadConfig loads furniture teleport policy from environment variables.
func LoadConfig() (Config, error) {
	return env.ParseAs[Config]()
}
