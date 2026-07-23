// Package subscription owns club membership and store offer behavior.
package subscription

import (
	"time"

	"github.com/caarlos0/env/v11"
)

// Config controls subscription scheduling and rewards.
type Config struct {
	// TickInterval stores scheduler frequency.
	TickInterval time.Duration `env:"PIXELS_SUBSCRIPTION_TICK_INTERVAL" envDefault:"1m"`
	// PaydayInterval stores the kickback cycle.
	PaydayInterval time.Duration `env:"PIXELS_SUBSCRIPTION_PAYDAY_INTERVAL" envDefault:"744h"`
	// KickbackPercentage stores the spending reward percentage.
	KickbackPercentage float64 `env:"PIXELS_SUBSCRIPTION_KICKBACK_PERCENTAGE" envDefault:"0.10"`
	// PaydayCurrencyType identifies the reward currency.
	PaydayCurrencyType int32 `env:"PIXELS_SUBSCRIPTION_PAYDAY_CURRENCY_TYPE" envDefault:"-1"`
	// BonusRareCurrencyType identifies the currency displayed by Bonus Rare.
	BonusRareCurrencyType int32 `env:"PIXELS_SUBSCRIPTION_BONUSRARE_CURRENCY_TYPE" envDefault:"5"`
	// BonusRareThreshold stores the balance required by Bonus Rare.
	BonusRareThreshold int64 `env:"PIXELS_SUBSCRIPTION_BONUSRARE_THRESHOLD" envDefault:"120"`
	// BonusRareProductID identifies the furniture definition shown as the reward.
	BonusRareProductID int32 `env:"PIXELS_SUBSCRIPTION_BONUSRARE_PRODUCT_ID" envDefault:"0"`
}

// LoadConfig loads subscription settings from environment variables.
func LoadConfig() (Config, error) {
	return env.ParseAs[Config]()
}

// Normalize fills invalid subscription values with conservative defaults.
func (config Config) Normalize() Config {
	if config.TickInterval <= 0 {
		config.TickInterval = time.Minute
	}
	if config.PaydayInterval <= 0 {
		config.PaydayInterval = 31 * 24 * time.Hour
	}
	if config.KickbackPercentage < 0 || config.KickbackPercentage > 1 {
		config.KickbackPercentage = 0.10
	}
	if config.BonusRareThreshold <= 0 {
		config.BonusRareThreshold = 120
	}
	if config.BonusRareProductID < 0 {
		config.BonusRareProductID = 0
	}
	return config
}
