package core

import "time"

// Options controls Marketplace pricing, limits, and scheduling.
type Options struct {
	// Enabled reports whether player Marketplace operations are available.
	Enabled bool
	// CommissionPercent stores the upward-rounded buyer surcharge percentage.
	CommissionPercent int64
	// TokenCost stores the credit cost of one token package.
	TokenCost int64
	// TokenPackageSize stores tokens granted per purchase.
	TokenPackageSize int32
	// AdvertisementCost stores Nitro's promoted-listing display cost.
	AdvertisementCost int64
	// MinimumPrice stores the minimum seller price.
	MinimumPrice int64
	// MaximumPrice stores the maximum seller price.
	MaximumPrice int64
	// OfferDuration stores open-listing lifetime.
	OfferDuration time.Duration
	// DisplayDuration stores historical listing display lifetime.
	DisplayDuration time.Duration
	// SearchCacheTTL stores shared search-cache lifetime.
	SearchCacheTTL time.Duration
	// ExpiryInterval stores expiry scheduler cadence.
	ExpiryInterval time.Duration
}
