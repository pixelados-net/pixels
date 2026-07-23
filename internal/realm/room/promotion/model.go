// Package promotion owns durable Room Ads and Room Event behavior.
package promotion

import "time"

// Promotion contains one room's current purchased event banner.
type Promotion struct {
	// ID identifies the promotion.
	ID int64
	// RoomID identifies the promoted room.
	RoomID int64
	// CategoryID identifies the event category.
	CategoryID int32
	// Title stores the visible event name.
	Title string
	// Description stores the visible event description.
	Description string
	// StartsAt stores the first activation time.
	StartsAt time.Time
	// EndsAt stores the active expiration boundary.
	EndsAt time.Time
	// CreatedBy identifies the purchasing player.
	CreatedBy int64
	// Version stores optimistic mutation order.
	Version int64
}

// ActiveAt reports whether the promotion is visible at a time.
func (promotion Promotion) ActiveAt(now time.Time) bool {
	return promotion.ID > 0 && promotion.EndsAt.After(now)
}

// PurchaseParams contains one requested Room Ad purchase.
type PurchaseParams struct {
	// PlayerID identifies the buyer.
	PlayerID int64
	// PlayerName stores the buyer's visible username.
	PlayerName string
	// RoomID identifies the promoted room.
	RoomID int64
	// PageID identifies the Room Ads catalog page.
	PageID int64
	// OfferID identifies the charged catalog offer.
	OfferID int64
	// Title stores visible event copy.
	Title string
	// Description stores visible event copy.
	Description string
	// CategoryID identifies the event category.
	CategoryID int32
	// Extended reports whether an active promotion should add duration.
	Extended bool
	// HasClub reports current catalog club eligibility.
	HasClub bool
}

// EditParams contains one promotion copy update.
type EditParams struct {
	// PlayerID identifies the editor.
	PlayerID int64
	// PromotionID identifies the event.
	PromotionID int64
	// Title replaces the visible event name.
	Title string
	// Description replaces the visible event description.
	Description string
}
