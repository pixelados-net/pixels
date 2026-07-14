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

	// RoomBundleTemplateRoomID identifies the room cloned by this offer.
	RoomBundleTemplateRoomID *int64

	// GrantsEffectID identifies an additional or effect-only reward.
	GrantsEffectID *int32

	// GrantsEffectDurationSeconds stores one granted charge duration.
	GrantsEffectDurationSeconds int32

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

	// BundleDiscountEnabled reports whether bulk discount purchases are allowed.
	BundleDiscountEnabled bool

	// Giftable reports whether another player may receive this offer as a gift.
	Giftable bool

	// ClubOnly reports whether the offer requires club membership.
	ClubOnly bool

	// OrderNum stores page display order.
	OrderNum int32

	// Enabled reports whether the offer can be purchased.
	Enabled bool

	// ExtraData stores the initial furniture protocol state.
	ExtraData string

	// ScheduledAt stores the optional future LTD publication time.
	ScheduledAt *time.Time
}

// IsLimited reports whether the offer belongs to an LTD series.
func (item Item) IsLimited() bool { return item.LimitedStack > 0 }

// IsCredits reports whether the offer uses credits.
func (item Item) IsCredits() bool { return item.PointsType == CreditsType }

// IsRoomBundle reports whether the offer creates a room from a template.
func (item Item) IsRoomBundle() bool { return item.RoomBundleTemplateRoomID != nil }

// BulkDiscountEligible reports whether amount greater than one is allowed.
func (item Item) BulkDiscountEligible(hasProducts bool) bool {
	return item.BundleDiscountEnabled && !item.IsLimited() && item.Amount == 1 && !hasProducts
}

// Product contains one furniture definition granted by a multi-product offer.
type Product struct {
	// ID identifies the durable product row.
	ID int64
	// CatalogItemID identifies the owning offer.
	CatalogItemID int64
	// DefinitionID identifies the granted furniture definition.
	DefinitionID int64
	// Quantity stores the number of granted instances.
	Quantity int32
	// OrderNum stores stable wire order.
	OrderNum int32
}

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
