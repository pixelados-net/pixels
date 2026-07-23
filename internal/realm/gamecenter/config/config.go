// Package config loads Game Center runtime policy.
package config

import (
	"os"
	"strconv"
)

// Config stores Game Center feature policy.
type Config struct {
	// Enabled controls all Game Center handlers.
	Enabled bool
}

// Load reads Game Center environment settings with safe defaults.
func Load() Config {
	value, err := strconv.ParseBool(os.Getenv("PIXELS_GAMECENTER_ENABLED"))
	if err != nil {
		value = true
	}
	return Config{Enabled: value}
}
