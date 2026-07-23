package marketplace

import (
	"testing"
	"time"
)

// TestLoadConfigNormalizesUnsafeValues verifies environment input cannot overflow Nitro prices.
func TestLoadConfigNormalizesUnsafeValues(t *testing.T) {
	t.Setenv("PIXELS_MARKETPLACE_COMMISSION_PERCENT", "999")
	t.Setenv("PIXELS_MARKETPLACE_TOKEN_COST", "0")
	t.Setenv("PIXELS_MARKETPLACE_TOKEN_PACKAGE_SIZE", "0")
	t.Setenv("PIXELS_MARKETPLACE_ADVERTISEMENT_COST", "-1")
	t.Setenv("PIXELS_MARKETPLACE_MIN_PRICE", "999999999999")
	t.Setenv("PIXELS_MARKETPLACE_MAX_PRICE", "999999999999")
	t.Setenv("PIXELS_MARKETPLACE_OFFER_DURATION", "0s")
	config := LoadConfig()
	if config.CommissionPercent != 100 || config.TokenCost != 1 || config.TokenPackageSize != 5 || config.AdvertisementCost != 0 || config.MinimumPrice != 1 || config.MaximumPrice <= 0 || config.OfferDuration != 48*time.Hour {
		t.Fatalf("config=%#v", config)
	}
	buyerPriceLimit := config.MaximumPrice + (config.MaximumPrice*config.CommissionPercent+99)/100
	if buyerPriceLimit > 2147483647 {
		t.Fatalf("buyer price limit=%d", buyerPriceLimit)
	}
}
