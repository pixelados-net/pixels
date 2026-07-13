// Package purchased contains the catalog purchase event.
package purchased

import "github.com/niflaot/pixels/pkg/bus"

const (
	// Name identifies one completed catalog purchase.
	Name bus.Name = "catalog.purchased"
)

// Payload contains one completed catalog purchase.
type Payload struct {
	// PlayerID identifies the buyer.
	PlayerID int64

	// CatalogItemID identifies the purchased offer.
	CatalogItemID int64

	// DefinitionID identifies the granted furniture definition.
	DefinitionID int64

	// Quantity stores the granted item count.
	Quantity int32

	// CostCredits stores the charged credits.
	CostCredits int64

	// CostPoints stores the charged activity points.
	CostPoints int64

	// PointsType identifies the charged activity-points currency.
	PointsType int32

	// LimitedUnitNumber stores the optional LTD edition number.
	LimitedUnitNumber *int32

	// CreatedRoomID identifies a room created by a room bundle offer.
	CreatedRoomID *int64
}
