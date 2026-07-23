package profile

import (
	"time"

	"github.com/caarlos0/env/v11"
)

// Config contains public-profile policy limits.
type Config struct {
	// MottoMaximumRunes bounds public mottos.
	MottoMaximumRunes int `env:"PIXELS_PLAYER_MOTTO_MAX_RUNES" envDefault:"38"`
	// TagMaximumCount bounds public tag replacements.
	TagMaximumCount int `env:"PIXELS_PLAYER_TAG_MAX_COUNT" envDefault:"5"`
	// TagMaximumRunes bounds each public tag.
	TagMaximumRunes int `env:"PIXELS_PLAYER_TAG_MAX_RUNES" envDefault:"32"`
	// DailyRespectLimit is the ordinary daily user-respect allowance.
	DailyRespectLimit int `env:"PIXELS_PLAYER_RESPECT_DAILY_LIMIT" envDefault:"3"`
	// DailyPetRespectLimit is the ordinary daily pet-respect allowance.
	DailyPetRespectLimit int `env:"PIXELS_PLAYER_PET_RESPECT_DAILY_LIMIT" envDefault:"3"`
	// RespectThrottle bounds repeated respect requests.
	RespectThrottle time.Duration `env:"PIXELS_PLAYER_RESPECT_THROTTLE" envDefault:"250ms"`
	// HotelTimezone defines the civil day used by daily respect allowances.
	HotelTimezone string `env:"PIXELS_HOTEL_TIMEZONE" envDefault:"America/Bogota"`
}

// LoadConfig loads public-profile policy from environment variables and defaults.
func LoadConfig() (Config, error) { return env.ParseAs[Config]() }

// DefaultConfig returns the documented public-profile policy defaults.
func DefaultConfig() Config {
	return Config{MottoMaximumRunes: 38, TagMaximumCount: MaxTags, TagMaximumRunes: MaxTagLength, DailyRespectLimit: DefaultDailyRespectLimit, DailyPetRespectLimit: DefaultPetRespectLimit, RespectThrottle: 250 * time.Millisecond, HotelTimezone: "America/Bogota"}
}
