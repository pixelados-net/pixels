// Package trade coordinates direct-trade capabilities and wiring.
package trade

import (
	tradecore "github.com/niflaot/pixels/internal/realm/trade/core"
	"os"
	"strconv"
	"time"
)

// Config aliases direct-trade core policy for dependency injection.
type Config = tradecore.Options

// LoadConfig loads direct-trade configuration with defaults.
func LoadConfig() Config {
	config := Config{Enabled: tradeBool("PIXELS_TRADE_ENABLED", true), StartThrottle: tradeDuration("PIXELS_TRADE_START_THROTTLE", 10*time.Second), MaximumItems: int(tradeInt("PIXELS_TRADE_MAX_ITEMS", 12)), AuditEnabled: tradeBool("PIXELS_TRADE_AUDIT_ENABLED", true)}
	if config.StartThrottle <= 0 {
		config.StartThrottle = 10 * time.Second
	}
	if config.MaximumItems < 1 {
		config.MaximumItems = 1
	}
	if config.MaximumItems > 100 {
		config.MaximumItems = 100
	}
	return config
}

// tradeBool loads one boolean.
func tradeBool(key string, fallback bool) bool {
	value, err := strconv.ParseBool(os.Getenv(key))
	if err != nil {
		return fallback
	}
	return value
}

// tradeDuration loads one duration.
func tradeDuration(key string, fallback time.Duration) time.Duration {
	value, err := time.ParseDuration(os.Getenv(key))
	if err != nil {
		return fallback
	}
	return value
}

// tradeInt loads one integer.
func tradeInt(key string, fallback int64) int64 {
	value, err := strconv.ParseInt(os.Getenv(key), 10, 64)
	if err != nil {
		return fallback
	}
	return value
}
