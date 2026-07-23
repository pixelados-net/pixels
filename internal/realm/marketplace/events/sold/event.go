// Package sold defines the marketplace.sold event.
package sold

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies a completed Marketplace purchase.
const Name bus.Name = "marketplace.sold"

// Payload describes a completed Marketplace purchase.
type Payload struct {
	// ListingID identifies the listing.
	ListingID int64
	// SellerPlayerID identifies the seller.
	SellerPlayerID int64
	// BuyerPlayerID identifies the buyer.
	BuyerPlayerID int64
	// FurnitureItemID identifies the transferred furniture.
	FurnitureItemID int64
	// RawPrice stores seller proceeds.
	RawPrice int64
	// BuyerPrice stores the commission-inclusive charge.
	BuyerPrice int64
}
