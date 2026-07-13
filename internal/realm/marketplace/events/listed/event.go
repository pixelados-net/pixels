// Package listed defines the marketplace.listed event.
package listed

import (
	"github.com/niflaot/pixels/pkg/bus"
	"time"
)

// Name identifies a newly opened listing.
const Name bus.Name = "marketplace.listed"

// Payload describes a newly opened listing.
type Payload struct {
	// ListingID identifies the listing.
	ListingID int64
	// SellerPlayerID identifies the seller.
	SellerPlayerID int64
	// FurnitureItemID identifies the reserved furniture instance.
	FurnitureItemID int64
	// FurnitureDefinitionID identifies the listed furniture type.
	FurnitureDefinitionID int64
	// RawPrice stores seller proceeds before commission.
	RawPrice int64
	// ExpiresAt stores the listing deadline.
	ExpiresAt time.Time
}
