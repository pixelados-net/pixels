// Package config contains dynamic plugin runtime settings.
package config

import (
	"strings"
	"time"

	"github.com/caarlos0/env/v11"
)

// Config controls plugin discovery, callback deadlines, and chat commands.
type Config struct {
	// Directory stores the root containing one folder per native plugin.
	Directory string `env:"PIXELS_PLUGIN_DIRECTORY" envDefault:"plugins"`
	// CallbackTimeout bounds interceptor, listener, and command callbacks.
	CallbackTimeout time.Duration `env:"PIXELS_PLUGIN_CALLBACK_TIMEOUT" envDefault:"2s"`
	// CommandPrefix identifies plugin command messages in room chat.
	CommandPrefix string `env:"PIXELS_COMMAND_PREFIX" envDefault:":"`
}

// Load reads plugin settings from environment variables.
func Load() (Config, error) {
	config, err := env.ParseAs[Config]()
	if err != nil {
		return Config{}, err
	}

	return config.Normalize(), nil
}

// Normalize fills invalid settings with conservative defaults.
func (config Config) Normalize() Config {
	config.Directory = strings.TrimSpace(config.Directory)
	if config.Directory == "" {
		config.Directory = "plugins"
	}
	if config.CallbackTimeout <= 0 {
		config.CallbackTimeout = 2 * time.Second
	}
	config.CommandPrefix = strings.TrimSpace(config.CommandPrefix)
	if config.CommandPrefix == "" {
		config.CommandPrefix = ":"
	}

	return config
}
