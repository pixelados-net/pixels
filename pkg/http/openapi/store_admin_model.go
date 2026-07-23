package openapi

import "time"

// VoucherRequest contains writable voucher data.
type VoucherRequest struct {
	APIKeyRequest
	// Code stores the voucher code.
	Code string `json:"code" required:"true"`
	// CostCredits stores granted credits.
	CostCredits int64 `json:"costCredits" minimum:"0"`
	// CostPoints stores granted activity points.
	CostPoints int64 `json:"costPoints" minimum:"0"`
	// PointsType identifies the granted activity-points currency.
	PointsType int32 `json:"pointsType"`
	// CatalogItemID identifies an optional free catalog reward.
	CatalogItemID *int64 `json:"catalogItemId,omitempty" minimum:"1"`
	// RedemptionCap optionally limits total voucher redemptions.
	RedemptionCap *int32 `json:"redemptionCap,omitempty" minimum:"1"`
	// PerPlayerCap stores the player redemption limit.
	PerPlayerCap int32 `json:"perPlayerCap" minimum:"1"`
	// Enabled reports whether redemption is active.
	Enabled bool `json:"enabled"`
	// ExpiresAt stores optional voucher expiration.
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

// VoucherPatchRequest contains a voucher id and writable data.
type VoucherPatchRequest struct {
	VoucherRequest
	// ID identifies the voucher.
	ID int64 `path:"id" required:"true" minimum:"1"`
}

// VoucherResponse contains one voucher record.
type VoucherResponse struct {
	// ID identifies the voucher.
	ID int64 `json:"id"`
	// Code stores the voucher code.
	Code string `json:"code"`
	// Enabled reports whether redemption is active.
	Enabled bool `json:"enabled"`
}

// VoucherListResponse contains voucher records.
type VoucherListResponse []VoucherResponse

// VoucherRedemptionResponse contains one voucher redemption.
type VoucherRedemptionResponse struct {
	// VoucherID identifies the redeemed voucher.
	VoucherID int64 `json:"voucherId"`
	// PlayerID identifies the player who redeemed the voucher.
	PlayerID int64 `json:"playerId"`
	// Redeemed stores the redemption timestamp.
	Redeemed time.Time `json:"redeemedAt"`
}

// VoucherRedemptionListResponse contains voucher redemption records.
type VoucherRedemptionListResponse []VoucherRedemptionResponse

// SubscriptionPlayerRequest identifies one player membership.
type SubscriptionPlayerRequest struct {
	APIKeyRequest
	// PlayerID identifies the player.
	PlayerID int64 `path:"playerId" required:"true" minimum:"1"`
}

// SubscriptionGrantRequest contains one membership grant.
type SubscriptionGrantRequest struct {
	SubscriptionPlayerRequest
	// Level stores HC or VIP tier.
	Level int16 `json:"level" minimum:"1" maximum:"2"`
	// DurationSeconds stores the extension duration.
	DurationSeconds int64 `json:"durationSeconds" minimum:"1"`
}

// SubscriptionIDRequest identifies one subscription configuration record.
type SubscriptionIDRequest struct {
	APIKeyRequest
	// ID identifies the record.
	ID int64 `path:"id" required:"true" minimum:"1"`
}

// SubscriptionResponse contains membership state and payday history.
type SubscriptionResponse struct {
	// Membership stores membership data.
	Membership map[string]any `json:"membership"`
	// PaydayProjection stores the current reward and countdown calculation.
	PaydayProjection map[string]any `json:"paydayProjection"`
	// GiftsAvailable stores currently claimable monthly rewards.
	GiftsAvailable int32 `json:"giftsAvailable"`
	// Paydays stores payday history.
	Paydays []map[string]any `json:"paydays"`
}

// ClubOfferRequest contains writable club offer data.
type ClubOfferRequest struct {
	APIKeyRequest
	// Name stores the stable product code.
	Name string `json:"name" required:"true"`
	// DayCount stores the membership duration.
	DayCount int32 `json:"dayCount" minimum:"1"`
	// PriceCredits stores the credits price.
	PriceCredits int64 `json:"priceCredits" minimum:"0"`
	// PricePoints stores the activity-points price.
	PricePoints int64 `json:"pricePoints" minimum:"0"`
	// PointsType identifies the activity-points currency.
	PointsType int32 `json:"pointsType"`
	// VIP reports whether this offer grants the VIP tier.
	VIP bool `json:"vip"`
	// Deal reports whether this is an extension deal.
	Deal bool `json:"deal"`
	// Enabled reports whether the offer is available.
	Enabled bool `json:"enabled"`
	// OrderNum stores display order.
	OrderNum int32 `json:"orderNum"`
}

// ClubOfferPatchRequest contains a club offer id and writable data.
type ClubOfferPatchRequest struct {
	ClubOfferRequest
	// ID identifies the club offer.
	ID int64 `path:"id" required:"true" minimum:"1"`
}

// ClubOfferResponse contains one club offer.
type ClubOfferResponse ClubOfferRequest

// ClubOfferListResponse contains club offers.
type ClubOfferListResponse []ClubOfferResponse

// TargetedOfferRequest contains writable personalized offer data.
type TargetedOfferRequest struct {
	APIKeyRequest
	// CatalogItemID identifies the purchased catalog offer.
	CatalogItemID int64 `json:"catalogItemId" minimum:"1"`
	// PriceCredits stores the credits override.
	PriceCredits int64 `json:"priceCredits" minimum:"0"`
	// PricePoints stores the activity-points override.
	PricePoints int64 `json:"pricePoints" minimum:"0"`
	// PointsType identifies the activity-points currency.
	PointsType int32 `json:"pointsType"`
	// PurchaseLimit stores the per-player limit.
	PurchaseLimit int32 `json:"purchaseLimit" minimum:"1"`
	// TitleKey stores the localized title key.
	TitleKey string `json:"titleKey" required:"true"`
	// DescriptionKey stores the localized description key.
	DescriptionKey string `json:"descriptionKey" required:"true"`
	// ImageURL stores banner artwork.
	ImageURL string `json:"imageUrl" required:"true"`
	// IconURL stores icon artwork.
	IconURL string `json:"iconUrl" required:"true"`
	// ExpiresAt stores the required future offer expiration.
	ExpiresAt *time.Time `json:"expiresAt" required:"true"`
	// OrderNum stores display order.
	OrderNum int32 `json:"orderNum"`
	// Enabled reports whether the offer is available.
	Enabled bool `json:"enabled"`
}

// TargetedOfferPatchRequest contains a targeted offer id and writable data.
type TargetedOfferPatchRequest struct {
	TargetedOfferRequest
	// ID identifies the targeted offer.
	ID int64 `path:"id" required:"true" minimum:"1"`
}

// TargetedOfferResponse contains one targeted offer.
type TargetedOfferResponse TargetedOfferRequest

// TargetedOfferListResponse contains targeted offers.
type TargetedOfferListResponse []TargetedOfferResponse

// CampaignRequest contains writable calendar campaign data.
type CampaignRequest struct {
	APIKeyRequest
	// Name stores the stable campaign name.
	Name string `json:"name" required:"true"`
	// Image stores campaign artwork.
	Image string `json:"image"`
	// StartDate stores campaign day zero.
	StartDate time.Time `json:"startDate" required:"true"`
	// DayCount stores the number of doors.
	DayCount int32 `json:"dayCount" minimum:"1"`
	// Enabled reports whether the campaign is active.
	Enabled bool `json:"enabled"`
	// Days stores optional campaign rewards.
	Days []CampaignDayRequest `json:"days"`
}

// CampaignDayRequest contains one calendar reward.
type CampaignDayRequest struct {
	// DayNumber identifies the zero-based door.
	DayNumber int32 `json:"dayNumber" minimum:"0"`
	// ProductDefinitionID optionally identifies granted furniture.
	ProductDefinitionID *int64 `json:"productDefinitionId,omitempty" minimum:"1"`
	// CustomImage stores reward artwork.
	CustomImage string `json:"customImage"`
	// CreditsReward stores granted credits.
	CreditsReward int64 `json:"creditsReward" minimum:"0"`
	// PointsReward stores granted activity points.
	PointsReward int64 `json:"pointsReward" minimum:"0"`
	// PointsType identifies the activity-points currency.
	PointsType int32 `json:"pointsType"`
}

// CampaignPatchRequest contains a campaign id and writable data.
type CampaignPatchRequest struct {
	CampaignRequest
	// ID identifies the campaign.
	ID int64 `path:"id" required:"true" minimum:"1"`
}

// CampaignResponse contains one calendar campaign.
type CampaignResponse CampaignRequest

// CampaignListResponse contains calendar campaigns.
type CampaignListResponse []CampaignResponse
