// Package record defines Marketplace persistence contracts and values.
package record

import "time"

// State describes the durable lifecycle of a listing.
type State int16

const (
	// StateOpen identifies a purchasable listing.
	StateOpen State = iota
	// StateSold identifies a purchased listing awaiting seller redemption.
	StateSold
	// StateClosed identifies a cancelled or expired listing.
	StateClosed
)

// Listing stores one durable Marketplace offer.
type Listing struct {
	// ID identifies the listing.
	ID int64
	// SellerPlayerID identifies the seller.
	SellerPlayerID int64
	// BuyerPlayerID optionally identifies the buyer.
	BuyerPlayerID *int64
	// FurnitureItemID identifies the unique furniture instance.
	FurnitureItemID int64
	// FurnitureDefinitionID identifies the searchable furniture type.
	FurnitureDefinitionID int64
	// RawPrice stores the seller price before commission.
	RawPrice int64
	// State stores the listing lifecycle.
	State State
	// ExpiresAt stores the open-listing deadline.
	ExpiresAt time.Time
	// SoldAt optionally stores purchase time.
	SoldAt *time.Time
	// RedeemedAt optionally stores seller redemption time.
	RedeemedAt *time.Time
	// CreatedAt stores creation time.
	CreatedAt time.Time
}

// DayStat stores one definition's daily sale aggregation.
type DayStat struct {
	// DefinitionID identifies the furniture type.
	DefinitionID int64
	// Day stores the UTC calendar day.
	Day time.Time
	// AverageRawPrice stores the mean seller price.
	AverageRawPrice int64
	// SoldCount stores completed sales.
	SoldCount int32
}

// Search contains bounded Marketplace filters.
type Search struct {
	// MinimumBuyerPrice stores the inclusive commission-adjusted price floor.
	MinimumBuyerPrice int64
	// MaximumBuyerPrice stores the inclusive commission-adjusted price ceiling.
	MaximumBuyerPrice int64
	// CommissionPercent stores the configured buyer surcharge.
	CommissionPercent int64
	// DefinitionIDs limits results to matching furniture definitions.
	DefinitionIDs []int64
	// SortType stores Nitro's requested ordering.
	SortType int32
	// Limit caps returned rows.
	Limit int32
}

// SearchOffer stores one aggregated Marketplace search group.
type SearchOffer struct {
	// Listing stores the cheapest representative listing in the group.
	Listing Listing
	// AverageRawPrice stores the rounded average seller price.
	AverageRawPrice int64
	// OfferCount stores the number of matching open listings.
	OfferCount int32
}
