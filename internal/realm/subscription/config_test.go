package subscription

import (
	"testing"
	"time"
)

// TestLoadConfigReadsSubscriptionEnvironment verifies configured scheduling values.
func TestLoadConfigReadsSubscriptionEnvironment(t *testing.T) {
	t.Setenv("PIXELS_SUBSCRIPTION_TICK_INTERVAL", "2s")
	t.Setenv("PIXELS_SUBSCRIPTION_PAYDAY_INTERVAL", "48h")
	t.Setenv("PIXELS_SUBSCRIPTION_KICKBACK_PERCENTAGE", "0.25")
	t.Setenv("PIXELS_SUBSCRIPTION_PAYDAY_CURRENCY_TYPE", "5")
	t.Setenv("PIXELS_SUBSCRIPTION_BONUSRARE_CURRENCY_TYPE", "6")
	t.Setenv("PIXELS_SUBSCRIPTION_BONUSRARE_THRESHOLD", "250")
	t.Setenv("PIXELS_SUBSCRIPTION_BONUSRARE_PRODUCT_ID", "9001")
	config, err := LoadConfig()
	if err != nil {
		t.Fatalf("load subscription config: %v", err)
	}
	if config.TickInterval != 2*time.Second || config.PaydayInterval != 48*time.Hour ||
		config.KickbackPercentage != 0.25 || config.PaydayCurrencyType != 5 || config.BonusRareCurrencyType != 6 ||
		config.BonusRareThreshold != 250 || config.BonusRareProductID != 9001 {
		t.Fatalf("unexpected config %#v", config)
	}
}

// TestNormalizeRestoresInvalidSubscriptionValues verifies conservative defaults.
func TestNormalizeRestoresInvalidSubscriptionValues(t *testing.T) {
	config := (Config{TickInterval: -1, PaydayInterval: -1, KickbackPercentage: 2, BonusRareThreshold: -1, BonusRareProductID: -1}).Normalize()
	if config.TickInterval != time.Minute || config.PaydayInterval != 31*24*time.Hour || config.KickbackPercentage != 0.1 ||
		config.BonusRareThreshold != 120 || config.BonusRareProductID != 0 {
		t.Fatalf("unexpected normalized config %#v", config)
	}
}
