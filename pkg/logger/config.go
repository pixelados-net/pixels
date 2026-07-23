// Package logger contains zap logger configuration and construction.
package logger

import "github.com/caarlos0/env/v11"

// Format names a zap encoder format.
type Format string

const (
	// FormatConsole writes human-readable console logs.
	FormatConsole Format = "console"

	// FormatJSON writes structured JSON logs.
	FormatJSON Format = "json"
)

// Config contains zap logger settings.
type Config struct {
	// Level is the minimum enabled zap level.
	Level string `env:"LOG_LEVEL" envDefault:"info"`

	// Format is the zap encoder format.
	Format Format `env:"LOG_FORMAT" envDefault:"console"`

	// ToonConsole enables compact agent-friendly console logs.
	ToonConsole bool `env:"TOON_CONSOLE" envDefault:"false"`
}

// LoadConfig reads logger configuration from environment variables.
func LoadConfig() (Config, error) {
	return env.ParseAs[Config]()
}
