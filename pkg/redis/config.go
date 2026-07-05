// Package redis contains reusable Redis storage components.
package redis

import "github.com/caarlos0/env/v11"

// Config contains Redis connection settings.
type Config struct {
	// Address is the Redis server address.
	Address string `env:"REDIS_ADDRESS" envDefault:"127.0.0.1:6379"`

	// Username is the Redis ACL username.
	Username string `env:"REDIS_USERNAME" envDefault:""`

	// Password is the Redis password.
	Password string `env:"REDIS_PASSWORD" envDefault:""`

	// Database is the selected Redis database.
	Database int `env:"REDIS_DATABASE" envDefault:"0"`
}

// LoadConfig reads Redis configuration from environment variables.
func LoadConfig() (Config, error) {
	return env.ParseAs[Config]()
}
