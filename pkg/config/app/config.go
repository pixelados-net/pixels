// Package app contains application-level configuration.
package app

import (
	"net"
	"strconv"

	"github.com/caarlos0/env/v11"
)

// Config contains application-level runtime settings.
type Config struct {
	// Environment names the runtime environment.
	Environment string `env:"PIXELS_ENV" envDefault:"development"`

	// Host is the protocol bind host.
	Host string `env:"PIXELS_HOST" envDefault:"127.0.0.1"`

	// Port is the protocol bind port.
	Port int `env:"PIXELS_PORT" envDefault:"3000"`

	// AccessKey protects private emulator endpoints.
	AccessKey string `env:"PIXELS_ACCESS_KEY" envDefault:"pixels-development-access-key-change-me"`
}

// Address returns the host and port formatted for a listener.
func (config Config) Address() string {
	return net.JoinHostPort(config.Host, strconv.Itoa(config.Port))
}

// Load reads application configuration from environment variables.
func Load() (Config, error) {
	return env.ParseAs[Config]()
}
