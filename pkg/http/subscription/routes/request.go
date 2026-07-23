package routes

import (
	"time"

	"github.com/niflaot/pixels/internal/realm/subscription/record"
)

// GrantRequest contains a manual membership grant.
type GrantRequest struct {
	// Level stores the granted club tier.
	Level record.Level `json:"level"`
	// DurationSeconds stores the extension duration.
	DurationSeconds int64 `json:"durationSeconds"`
}

// ClubOfferRequest contains writable club offer fields.
type ClubOfferRequest struct {
	// Name stores the product code.
	Name string `json:"name"`
	// DayCount stores granted days.
	DayCount int32 `json:"dayCount"`
	// PriceCredits stores the credits price.
	PriceCredits int64 `json:"priceCredits"`
	// PricePoints stores the points price.
	PricePoints int64 `json:"pricePoints"`
	// PointsType identifies the points currency.
	PointsType int32 `json:"pointsType"`
	// VIP reports whether the offer grants VIP.
	VIP bool `json:"vip"`
	// Deal reports whether the offer is extension-only.
	Deal bool `json:"deal"`
	// Enabled reports whether the offer is available.
	Enabled bool `json:"enabled"`
	// OrderNum stores display order.
	OrderNum int32 `json:"orderNum"`
}

// TargetedOfferRequest contains writable targeted-offer fields.
type TargetedOfferRequest struct {
	// CatalogItemID identifies the catalog offer.
	CatalogItemID int64 `json:"catalogItemId"`
	// PriceCredits stores the credits override.
	PriceCredits int64 `json:"priceCredits"`
	// PricePoints stores the points override.
	PricePoints int64 `json:"pricePoints"`
	// PointsType identifies the points currency.
	PointsType int32 `json:"pointsType"`
	// PurchaseLimit stores the player limit.
	PurchaseLimit int32 `json:"purchaseLimit"`
	// TitleKey stores the localized title key.
	TitleKey string `json:"titleKey"`
	// DescriptionKey stores the localized description key.
	DescriptionKey string `json:"descriptionKey"`
	// ImageURL stores banner artwork.
	ImageURL string `json:"imageUrl"`
	// IconURL stores icon artwork.
	IconURL string `json:"iconUrl"`
	// ExpiresAt stores the required future expiration.
	ExpiresAt *time.Time `json:"expiresAt"`
	// OrderNum stores display order.
	OrderNum int32 `json:"orderNum"`
	// Enabled reports whether the targeted offer is available.
	Enabled bool `json:"enabled"`
}

// CampaignRequest contains writable campaign data.
type CampaignRequest struct {
	// Name stores the stable campaign name.
	Name string `json:"name"`
	// Image stores campaign artwork.
	Image string `json:"image"`
	// StartDate stores campaign day zero.
	StartDate time.Time `json:"startDate"`
	// DayCount stores the number of doors.
	DayCount int32 `json:"dayCount"`
	// Enabled reports whether the campaign is active.
	Enabled bool `json:"enabled"`
	// Days stores optional campaign rewards.
	Days []record.CampaignDay `json:"days"`
}
