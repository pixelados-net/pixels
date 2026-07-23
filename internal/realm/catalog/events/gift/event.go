// Package gift contains the catalog gift purchase event.
package gift

import "github.com/niflaot/pixels/pkg/bus"

const (
	// Name identifies one committed catalog gift purchase.
	Name bus.Name = "catalog.purchased.gift"
)

// Payload contains one catalog gift purchase.
type Payload struct {
	// BuyerID identifies the paying player.
	BuyerID int64
	// ReceiverID identifies the receiving player.
	ReceiverID int64
	// CatalogItemID identifies the purchased offer.
	CatalogItemID int64
}
