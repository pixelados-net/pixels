// Package bundlepurchased contains the completed room bundle purchase event.
package bundlepurchased

import "github.com/niflaot/pixels/pkg/bus"

const (
	// Name identifies a completed room bundle purchase.
	Name bus.Name = "room.bundle.purchased"
)

// Payload contains one completed room bundle purchase.
type Payload struct {
	// PlayerID identifies the buyer.
	PlayerID int64
	// CatalogItemID identifies the purchased offer.
	CatalogItemID int64
	// TemplateRoomID identifies the cloned template.
	TemplateRoomID int64
	// CreatedRoomID identifies the new room.
	CreatedRoomID int64
	// FurnitureCount stores copied furniture count.
	FurnitureCount int
	// BotCount stores copied bot count.
	BotCount int
}
