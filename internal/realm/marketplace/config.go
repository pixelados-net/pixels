// Package marketplace coordinates Marketplace capabilities and wiring.
package marketplace

import (
	"math"
	"os"
	"strconv"
	"time"

	marketcore "github.com/niflaot/pixels/internal/realm/marketplace/core"
)

// Config aliases Marketplace core policy for dependency injection.
type Config = marketcore.Options

// LoadConfig loads Marketplace configuration with safe defaults.
func LoadConfig() Config {
	packageSize := envInt64("PIXELS_MARKETPLACE_TOKEN_PACKAGE_SIZE", 5)
	if packageSize < 1 || packageSize > math.MaxInt32 {
		packageSize = 5
	}
	config := Config{Enabled: envBool("PIXELS_MARKETPLACE_ENABLED", true), CommissionPercent: envInt64("PIXELS_MARKETPLACE_COMMISSION_PERCENT", 1), TokenCost: envInt64("PIXELS_MARKETPLACE_TOKEN_COST", 1), TokenPackageSize: int32(packageSize), AdvertisementCost: envInt64("PIXELS_MARKETPLACE_ADVERTISEMENT_COST", 0), MinimumPrice: envInt64("PIXELS_MARKETPLACE_MIN_PRICE", 1), MaximumPrice: envInt64("PIXELS_MARKETPLACE_MAX_PRICE", 1000000), OfferDuration: envDuration("PIXELS_MARKETPLACE_OFFER_DURATION", 48*time.Hour), DisplayDuration: envDuration("PIXELS_MARKETPLACE_DISPLAY_DURATION", 7*24*time.Hour), SearchCacheTTL: envDuration("PIXELS_MARKETPLACE_SEARCH_CACHE_TTL", 30*time.Second), ExpiryInterval: envDuration("PIXELS_MARKETPLACE_EXPIRY_INTERVAL", time.Minute)}
	if config.CommissionPercent < 0 {
		config.CommissionPercent = 0
	}
	if config.CommissionPercent > 100 {
		config.CommissionPercent = 100
	}
	if config.TokenCost < 1 || config.TokenCost > math.MaxInt32 {
		config.TokenCost = 1
	}
	if config.AdvertisementCost < 0 || config.AdvertisementCost > math.MaxInt32 {
		config.AdvertisementCost = 0
	}
	maximumRawPrice := int64(math.MaxInt32) * 100 / (100 + config.CommissionPercent)
	if config.MinimumPrice < 1 || config.MinimumPrice > maximumRawPrice {
		config.MinimumPrice = 1
	}
	if config.MaximumPrice < config.MinimumPrice || config.MaximumPrice > maximumRawPrice {
		config.MaximumPrice = maximumRawPrice
	}
	if config.OfferDuration <= 0 {
		config.OfferDuration = 48 * time.Hour
	}
	if config.DisplayDuration <= 0 {
		config.DisplayDuration = 7 * 24 * time.Hour
	}
	if config.SearchCacheTTL <= 0 {
		config.SearchCacheTTL = 30 * time.Second
	}
	if config.ExpiryInterval <= 0 {
		config.ExpiryInterval = time.Minute
	}
	return config
}

// envInt64 loads one integer or its default.
func envInt64(key string, fallback int64) int64 {
	value, err := strconv.ParseInt(os.Getenv(key), 10, 64)
	if err != nil {
		return fallback
	}
	return value
}

// envBool loads one boolean or its default.
func envBool(key string, fallback bool) bool {
	value, err := strconv.ParseBool(os.Getenv(key))
	if err != nil {
		return fallback
	}
	return value
}

// envDuration loads one duration or its default.
func envDuration(key string, fallback time.Duration) time.Duration {
	value, err := time.ParseDuration(os.Getenv(key))
	if err != nil {
		return fallback
	}
	return value
}
