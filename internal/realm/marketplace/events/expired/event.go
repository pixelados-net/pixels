// Package expired defines the marketplace.expired event.
package expired

import "github.com/niflaot/pixels/pkg/bus"

// Name identifies an expired Marketplace listing.
const Name bus.Name = "marketplace.expired"

// Payload describes one expired listing.
type Payload struct {
	// ListingID identifies the expired listing.
	ListingID int64
	// SellerPlayerID identifies the seller.
	SellerPlayerID int64
	// FurnitureItemID identifies the returned furniture.
	FurnitureItemID int64
}
