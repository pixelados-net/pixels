package record

import "time"

// TargetedOffer contains one personalized catalog offer.
type TargetedOffer struct {
	// ID identifies the offer.
	ID int64
	// CatalogItemID identifies the granted catalog item.
	CatalogItemID int64
	// PriceCredits stores the override credits price.
	PriceCredits int64
	// PricePoints stores the override points price.
	PricePoints int64
	// PointsType identifies the points currency.
	PointsType int32
	// PurchaseLimit stores the per-player limit.
	PurchaseLimit int32
	// TitleKey stores the localized title key.
	TitleKey string
	// DescriptionKey stores the localized description key.
	DescriptionKey string
	// ImageURL stores the main image.
	ImageURL string
	// IconURL stores the icon image.
	IconURL string
	// ExpiresAt stores optional expiration.
	ExpiresAt *time.Time
	// OrderNum stores selection order.
	OrderNum int32
	// Enabled reports whether the offer may be selected.
	Enabled bool
	// PurchasesCount stores player progress.
	PurchasesCount int32
	// Dismissed reports player dismissal.
	Dismissed bool
}

// Campaign contains one seasonal calendar.
type Campaign struct {
	// ID identifies the campaign.
	ID int64
	// Name stores its stable name.
	Name string
	// Image stores the client image.
	Image string
	// StartDate stores calendar day zero.
	StartDate time.Time
	// DayCount stores the number of doors.
	DayCount int32
	// Enabled reports whether the campaign is active.
	Enabled bool
}

// CampaignDay contains one calendar reward.
type CampaignDay struct {
	// CampaignID identifies the campaign.
	CampaignID int64
	// DayNumber identifies the zero-based door.
	DayNumber int32
	// ProductDefinitionID optionally identifies furniture.
	ProductDefinitionID *int64
	// CustomImage stores reward artwork.
	CustomImage string
	// CreditsReward stores granted credits.
	CreditsReward int64
	// PointsReward stores granted activity points.
	PointsReward int64
	// PointsType identifies the activity-points currency.
	PointsType int32
}

// SeasonalOffer links one date to a catalog offer.
type SeasonalOffer struct {
	// Date identifies the active date.
	Date time.Time
	// CatalogPageID identifies the page.
	CatalogPageID int64
	// CatalogItemID identifies the offer.
	CatalogItemID int64
}
