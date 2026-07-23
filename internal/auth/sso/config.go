// Package sso creates and consumes one-time SSO tickets.
package sso

import (
	"time"

	"github.com/caarlos0/env/v11"
)

// Config contains SSO ticket settings.
type Config struct {
	// DefaultTTL is the default ticket lifetime.
	DefaultTTL time.Duration `env:"SSO_DEFAULT_TTL" envDefault:"5m"`

	// Key signs ticket storage keys before writing them to Redis.
	Key string `env:"SSO_KEY" envDefault:"pixels-development-sso-key-change-me"`

	// Prefix is the Redis key prefix for SSO tickets.
	Prefix string `env:"SSO_PREFIX" envDefault:"pixels:sso"`
}

// LoadConfig reads SSO configuration from environment variables.
func LoadConfig() (Config, error) {
	return env.ParseAs[Config]()
}
