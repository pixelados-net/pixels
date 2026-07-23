// Package config loads progression runtime policy.
package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config stores achievement, quest, talent, and quiz runtime policy.
type Config struct {
	// Enabled controls progression behavior.
	Enabled bool
	// TradeRequiresPerk gates trading behind the TRADE perk.
	TradeRequiresPerk bool
	// GuideMinimumTrackLevel gates guide duty behind the helper track.
	GuideMinimumTrackLevel int32
	// PresenceInterval controls online-presence progress cadence.
	PresenceInterval time.Duration
	// FlushInterval controls deferred progress persistence cadence.
	FlushInterval time.Duration
	// DailyPoolSeed stabilizes deterministic daily quest selection.
	DailyPoolSeed string
}

// Load reads progression environment settings with safe defaults.
func Load() Config {
	return Config{
		Enabled:                envBool("PIXELS_PROGRESSION_ENABLED", true),
		TradeRequiresPerk:      envBool("PIXELS_PROGRESSION_TRADE_REQUIRES_PERK", false),
		GuideMinimumTrackLevel: int32(envInt("PIXELS_PROGRESSION_GUIDE_MIN_TRACK_LEVEL", 0)),
		PresenceInterval:       envDuration("PIXELS_PROGRESSION_PRESENCE_INTERVAL", 5*time.Minute),
		FlushInterval:          envDuration("PIXELS_PROGRESSION_FLUSH_INTERVAL", 2*time.Second),
		DailyPoolSeed:          strings.TrimSpace(os.Getenv("PIXELS_PROGRESSION_DAILY_POOL_SEED")),
	}
}

// envBool reads one optional boolean environment value.
func envBool(name string, fallback bool) bool {
	value, err := strconv.ParseBool(os.Getenv(name))
	if err != nil {
		return fallback
	}
	return value
}

// envInt reads one optional non-negative integer environment value.
func envInt(name string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(name))
	if err != nil || value < 0 {
		return fallback
	}
	return value
}

// envDuration reads one optional positive duration environment value.
func envDuration(name string, fallback time.Duration) time.Duration {
	value, err := time.ParseDuration(os.Getenv(name))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}
