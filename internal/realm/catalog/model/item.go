package model

import (
	"time"

	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

const (
	// CreditsType identifies the protocol credits balance.
	CreditsType int32 = -1
)

// Item contains one persistent catalog offer.
type Item struct {
	// Base contains shared durable record fields.
	sharedmodel.Base

	// PageID identifies the owning catalog page.
	PageID int64

	// DefinitionID identifies the granted furniture definition.
	DefinitionID int64

	// Name stores the stable localization slug.
	Name string

	// CostCredits stores the credits price.
	CostCredits int64

	// CostPoints stores the activity-points price.
	CostPoints int64

	// PointsType identifies the activity-points currency.
	PointsType int32

	// Amount stores the number of furniture instances granted.
	Amount int32

	// LimitedStack stores the LTD series size, or zero for regular offers.
	LimitedStack int32

	// LimitedSells stores the committed LTD sale count.
	LimitedSells int32

	// OfferID optionally groups catalog rows under one client offer.
	OfferID *int64

	// ClubOnly reports whether the offer requires club membership.
	ClubOnly bool

	// OrderNum stores page display order.
	OrderNum int32

	// Enabled reports whether the offer can be purchased.
	Enabled bool

	// ExtraData stores the initial furniture protocol state.
	ExtraData string
}

// IsLimited reports whether the offer belongs to an LTD series.
func (item Item) IsLimited() bool { return item.LimitedStack > 0 }

// IsCredits reports whether the offer uses credits.
func (item Item) IsCredits() bool { return item.PointsType == CreditsType }

// LimitedUnit contains one numbered LTD allocation.
type LimitedUnit struct {
	// ID identifies the durable allocation.
	ID int64

	// CatalogItemID identifies the LTD offer.
	CatalogItemID int64

	// UnitNumber stores the edition number.
	UnitNumber int32

	// OwnerPlayerID identifies the buyer after reservation.
	OwnerPlayerID *int64

	// FurnitureItemID identifies the granted instance after completion.
	FurnitureItemID *int64

	// SoldAt stores the reservation timestamp.
	SoldAt *time.Time
}
