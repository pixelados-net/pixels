// Package entry decides whether a player may enter a room.
package entry

import (
	"time"

	"github.com/caarlos0/env/v11"
	"golang.org/x/crypto/bcrypt"
)

const (
	// defaultLockoutSeconds stores the default ten-minute lockout.
	defaultLockoutSeconds int64 = 600
	// maxLockoutSeconds stores the largest value representable by time.Duration.
	maxLockoutSeconds int64 = (1<<63 - 1) / int64(time.Second)
)

// Config controls closed-room entry behavior.
type Config struct {
	// HangoutTimeout stores how long a doorbell request may wait.
	HangoutTimeout time.Duration `env:"PIXELS_ROOM_ENTRY_HANGOUT_TIMEOUT" envDefault:"5m"`
	// MaxPasswordAttempts stores failed attempts allowed before lockout.
	MaxPasswordAttempts int64 `env:"PIXELS_ROOM_ENTRY_MAX_PASSWORD_ATTEMPTS" envDefault:"5"`
	// AttemptWindow stores how long failed password attempts accumulate.
	AttemptWindow time.Duration `env:"PIXELS_ROOM_ENTRY_ATTEMPT_WINDOW" envDefault:"5m"`
	// LockoutSeconds stores how many seconds password entry remains frozen.
	LockoutSeconds int64 `env:"PIXELS_ROOM_ENTRY_LOCKOUT_SECONDS" envDefault:"600"`
	// PasswordCost stores the bcrypt cost used when room passwords are created.
	PasswordCost int `env:"PIXELS_ROOM_ENTRY_PASSWORD_COST" envDefault:"10"`
	// TrustedTTL stores how long an admin entry bypass remains usable.
	TrustedTTL time.Duration `env:"PIXELS_ROOM_ENTRY_TRUSTED_TTL" envDefault:"10s"`
}

// LoadConfig reads room entry configuration from environment variables.
func LoadConfig() (Config, error) {
	return env.ParseAs[Config]()
}

// Normalize fills invalid values with conservative defaults.
func (config Config) Normalize() Config {
	if config.HangoutTimeout <= 0 {
		config.HangoutTimeout = 5 * time.Minute
	}
	if config.MaxPasswordAttempts <= 0 {
		config.MaxPasswordAttempts = 5
	}
	if config.AttemptWindow <= 0 {
		config.AttemptWindow = 5 * time.Minute
	}
	if config.LockoutSeconds <= 0 {
		config.LockoutSeconds = defaultLockoutSeconds
	}
	if config.LockoutSeconds > maxLockoutSeconds {
		config.LockoutSeconds = maxLockoutSeconds
	}
	if config.PasswordCost < bcrypt.MinCost || config.PasswordCost > bcrypt.MaxCost {
		config.PasswordCost = bcrypt.DefaultCost
	}
	if config.TrustedTTL <= 0 {
		config.TrustedTTL = 10 * time.Second
	}

	return config
}

// LockoutDuration returns the configured password lockout as a Go duration.
func (config Config) LockoutDuration() time.Duration {
	return time.Duration(config.Normalize().LockoutSeconds) * time.Second
}
