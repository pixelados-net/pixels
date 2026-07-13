// Package core implements Marketplace business workflows.
package core

import (
	"time"

	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	marketrecord "github.com/niflaot/pixels/internal/realm/marketplace/record"
)

// Offer joins a listing with its immutable item and definition projection.
type Offer struct {
	// Listing stores durable sale state.
	Listing marketrecord.Listing
	// Item stores the furniture instance.
	Item furnituremodel.Item
	// Definition stores furniture metadata.
	Definition furnituremodel.Definition
	// BuyerPrice stores the commission-inclusive price.
	BuyerPrice int64
	// MinutesRemaining stores non-negative open lifetime.
	MinutesRemaining int32
	// AveragePrice stores the commission-adjusted group average.
	AveragePrice int64
	// OfferCount stores matching open listings in the search group.
	OfferCount int32
}

// SearchParams contains Nitro Marketplace search input.
type SearchParams struct {
	// MinimumPrice stores inclusive buyer price.
	MinimumPrice int64
	// MaximumPrice stores inclusive buyer price.
	MaximumPrice int64
	// Query stores a case-insensitive definition search.
	Query string
	// SortType stores Nitro ordering.
	SortType int32
}

// SearchResult contains bounded Marketplace results.
type SearchResult struct {
	// Offers stores at most 250 records.
	Offers []Offer
	// Total stores the total matching result count.
	Total int32
}

// Stats contains current and historical sale information.
type Stats struct {
	// AveragePrice stores the latest buyer-price average.
	AveragePrice int64
	// OpenCount stores current open listings.
	OpenCount int32
	// History stores recent daily stats.
	History []marketrecord.DayStat
}

// nowMinutes returns a bounded listing lifetime.
func nowMinutes(expires time.Time, now time.Time) int32 {
	if !expires.After(now) {
		return 0
	}
	minutes := expires.Sub(now) / time.Minute
	if minutes > 2147483647 {
		return 2147483647
	}
	return int32(minutes)
}
