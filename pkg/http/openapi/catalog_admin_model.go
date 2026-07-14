package openapi

// CatalogPageRequest contains one catalog page mutation.
type CatalogPageRequest struct {
	APIKeyRequest
	// ParentID identifies the optional parent page.
	ParentID *int64 `json:"parentId,omitempty" minimum:"1"`
	// Name stores the stable localization slug.
	Name string `json:"name" required:"true" example:"chairs"`
	// Layout identifies the Nitro catalog layout.
	Layout string `json:"layout" required:"true" example:"default_3x3"`
	// IconColor stores the client icon color.
	IconColor int32 `json:"iconColor" minimum:"0"`
	// IconImage stores the client icon image.
	IconImage int32 `json:"iconImage" minimum:"0"`
	// RequiredNode stores the optional permission needed to access the page.
	RequiredNode *string `json:"requiredNode,omitempty" example:"catalog.page.staff"`
	// OrderNum stores sibling display order.
	OrderNum int32 `json:"orderNum"`
	// Visible reports whether the page appears in the tree.
	Visible bool `json:"visible"`
	// Enabled reports whether the page can be opened.
	Enabled bool `json:"enabled"`
	// ClubOnly reports whether club membership is required.
	ClubOnly bool `json:"clubOnly"`
}

// CatalogPagePatchRequest contains optional catalog page changes.
type CatalogPagePatchRequest struct {
	APIKeyRequest
	// ID identifies the mutated page.
	ID int64 `path:"id" required:"true" minimum:"1"`
	// ParentID replaces the optional parent page.
	ParentID *int64 `json:"parentId,omitempty" minimum:"1"`
	// Name replaces the localization slug.
	Name *string `json:"name,omitempty"`
	// Layout replaces the Nitro layout.
	Layout *string `json:"layout,omitempty"`
	// IconColor replaces the icon color.
	IconColor *int32 `json:"iconColor,omitempty" minimum:"0"`
	// IconImage replaces the icon image.
	IconImage *int32 `json:"iconImage,omitempty" minimum:"0"`
	// RequiredNode replaces the page access permission.
	RequiredNode *string `json:"requiredNode,omitempty" example:"catalog.page.staff"`
	// ClearRequiredNode removes the page access permission.
	ClearRequiredNode bool `json:"clearRequiredNode,omitempty"`
	// OrderNum replaces sibling display order.
	OrderNum *int32 `json:"orderNum,omitempty"`
	// Visible replaces page tree visibility.
	Visible *bool `json:"visible,omitempty"`
	// Enabled replaces page availability.
	Enabled *bool `json:"enabled,omitempty"`
	// ClubOnly replaces club access policy.
	ClubOnly *bool `json:"clubOnly,omitempty"`
}

// CatalogPageResponse contains one catalog page record.
type CatalogPageResponse struct {
	// ID identifies the page.
	ID int64 `json:"id" required:"true"`
	// ParentID identifies the optional parent page.
	ParentID *int64 `json:"parentId"`
	// Name stores the localization slug.
	Name string `json:"name" required:"true"`
	// Layout identifies the Nitro layout.
	Layout string `json:"layout" required:"true"`
	// RequiredNode stores the optional page access permission.
	RequiredNode *string `json:"requiredNode,omitempty"`
	// Visible reports tree visibility.
	Visible bool `json:"visible" required:"true"`
	// Enabled reports page availability.
	Enabled bool `json:"enabled" required:"true"`
	// Version stores optimistic locking state.
	Version int64 `json:"version" required:"true"`
}

// CatalogPagesResponse contains catalog page records.
type CatalogPagesResponse []CatalogPageResponse

// CatalogItemRequest contains one catalog offer mutation.
type CatalogItemRequest struct {
	APIKeyRequest
	// PageID identifies the owning page.
	PageID int64 `json:"pageId" required:"true" minimum:"1"`
	// DefinitionID identifies the granted furniture definition.
	DefinitionID int64 `json:"definitionId,omitempty" minimum:"1"`
	// RoomBundleTemplateRoomID identifies a marked room template instead of furniture.
	RoomBundleTemplateRoomID *int64 `json:"roomBundleTemplateRoomId,omitempty" minimum:"1"`
	// GrantsEffectID identifies an additional or effect-only reward.
	GrantsEffectID *int32 `json:"grantsEffectId,omitempty" minimum:"1"`
	// GrantsEffectDurationSeconds stores one charge duration.
	GrantsEffectDurationSeconds int32 `json:"grantsEffectDurationSeconds" minimum:"0"`
	// Name stores the localization slug.
	Name string `json:"name" required:"true" example:"chair_plasto"`
	// CostCredits stores the credits price.
	CostCredits int64 `json:"costCredits" minimum:"0"`
	// CostPoints stores the activity-points price.
	CostPoints int64 `json:"costPoints" minimum:"0"`
	// PointsType identifies the activity-points currency or -1 for credits.
	PointsType int32 `json:"pointsType" required:"true" example:"-1"`
	// Amount stores the furniture quantity granted.
	Amount int32 `json:"amount" minimum:"0"`
	// LimitedStack stores numbered stock or zero.
	LimitedStack int32 `json:"limitedStack" minimum:"0"`
	// BundleDiscountEnabled permits bulk purchases with protocol discounts.
	BundleDiscountEnabled bool `json:"bundleDiscountEnabled"`
	// Giftable permits purchasing the offer for another player.
	Giftable bool `json:"giftable"`
	// ClubOnly reports club access policy.
	ClubOnly bool `json:"clubOnly"`
	// OrderNum stores page display order.
	OrderNum int32 `json:"orderNum"`
	// Enabled reports offer availability.
	Enabled bool `json:"enabled"`
	// ExtraData stores initial furniture state.
	ExtraData string `json:"extraData" example:"0"`
}

// CatalogItemPatchRequest contains optional catalog offer changes.
type CatalogItemPatchRequest struct {
	APIKeyRequest
	// ID identifies the mutated offer.
	ID int64 `path:"id" required:"true" minimum:"1"`
	// PageID replaces the owning page.
	PageID *int64 `json:"pageId,omitempty" minimum:"1"`
	// CostCredits replaces the credits price.
	CostCredits *int64 `json:"costCredits,omitempty" minimum:"0"`
	// CostPoints replaces the points price.
	CostPoints *int64 `json:"costPoints,omitempty" minimum:"0"`
	// LimitedStack replaces numbered stock without deleting completed sales.
	LimitedStack *int32 `json:"limitedStack,omitempty" minimum:"0"`
	// Enabled replaces offer availability.
	Enabled *bool `json:"enabled,omitempty"`
	// Name replaces the localization slug.
	Name *string `json:"name,omitempty"`
	// DefinitionID replaces the furniture definition.
	DefinitionID *int64 `json:"definitionId,omitempty" minimum:"1"`
	// RoomBundleTemplateRoomID replaces the marked room template.
	RoomBundleTemplateRoomID *int64 `json:"roomBundleTemplateRoomId,omitempty" minimum:"1"`
	// ClearRoomBundleTemplate removes the room template association.
	ClearRoomBundleTemplate bool `json:"clearRoomBundleTemplate,omitempty"`
	// GrantsEffectID replaces the effect reward.
	GrantsEffectID *int32 `json:"grantsEffectId,omitempty" minimum:"1"`
	// ClearGrantsEffect removes the effect reward.
	ClearGrantsEffect bool `json:"clearGrantsEffect,omitempty"`
	// GrantsEffectDurationSeconds replaces one charge duration.
	GrantsEffectDurationSeconds *int32 `json:"grantsEffectDurationSeconds,omitempty" minimum:"0"`
	// PointsType replaces the points currency or -1 for credits.
	PointsType *int32 `json:"pointsType,omitempty"`
	// Amount replaces the granted furniture amount.
	Amount *int32 `json:"amount,omitempty" minimum:"0"`
	// BundleDiscountEnabled replaces bulk discount eligibility.
	BundleDiscountEnabled *bool `json:"bundleDiscountEnabled,omitempty"`
	// Giftable replaces gift eligibility.
	Giftable *bool `json:"giftable,omitempty"`
	// ClubOnly replaces club access policy.
	ClubOnly *bool `json:"clubOnly,omitempty"`
	// OrderNum replaces page display order.
	OrderNum *int32 `json:"orderNum,omitempty"`
	// ExtraData replaces initial furniture state.
	ExtraData *string `json:"extraData,omitempty"`
}

// CatalogItemResponse contains one catalog offer record.
type CatalogItemResponse struct {
	// ID identifies the offer.
	ID int64 `json:"id" required:"true"`
	// PageID identifies the owning page.
	PageID int64 `json:"pageId" required:"true"`
	// DefinitionID identifies the furniture definition.
	DefinitionID int64 `json:"definitionId" required:"true"`
	// RoomBundleTemplateRoomID identifies the cloned room template.
	RoomBundleTemplateRoomID *int64 `json:"roomBundleTemplateRoomId,omitempty"`
	// GrantsEffectID identifies the effect reward.
	GrantsEffectID *int32 `json:"grantsEffectId,omitempty"`
	// GrantsEffectDurationSeconds stores one charge duration.
	GrantsEffectDurationSeconds int32 `json:"grantsEffectDurationSeconds" required:"true"`
	// Name stores the localization slug.
	Name string `json:"name" required:"true"`
	// CostCredits stores the credits price.
	CostCredits int64 `json:"costCredits" required:"true"`
	// CostPoints stores the points price.
	CostPoints int64 `json:"costPoints" required:"true"`
	// PointsType identifies the points currency.
	PointsType int32 `json:"pointsType" required:"true"`
	// Amount stores the granted furniture amount, or zero for room bundles.
	Amount int32 `json:"amount" required:"true"`
	// LimitedStack stores numbered stock.
	LimitedStack int32 `json:"limitedStack" required:"true"`
	// LimitedSells stores committed numbered sales.
	LimitedSells int32 `json:"limitedSells" required:"true"`
	// BundleDiscountEnabled reports bulk purchase eligibility.
	BundleDiscountEnabled bool `json:"bundleDiscountEnabled" required:"true"`
	// Giftable reports gift eligibility.
	Giftable bool `json:"giftable" required:"true"`
	// ClubOnly reports club access policy.
	ClubOnly bool `json:"clubOnly" required:"true"`
	// OrderNum stores page display order.
	OrderNum int32 `json:"orderNum" required:"true"`
	// Enabled reports offer availability.
	Enabled bool `json:"enabled" required:"true"`
	// ExtraData stores initial furniture state.
	ExtraData string `json:"extraData" required:"true"`
	// Version stores optimistic locking state.
	Version int64 `json:"version" required:"true"`
}

// CatalogItemsRequest contains optional catalog offer filters.
type CatalogItemsRequest struct {
	APIKeyRequest
	// PageID restricts offers to one page.
	PageID *int64 `query:"pageId,omitempty" minimum:"1"`
}

// CatalogItemsResponse contains catalog offer records.
type CatalogItemsResponse []CatalogItemResponse

// CatalogIDRequest identifies one catalog record.
type CatalogIDRequest struct {
	APIKeyRequest
	// ID identifies the catalog record.
	ID int64 `path:"id" required:"true" minimum:"1"`
}

// CatalogDefinitionResponse contains one definition missing an active offer.
type CatalogDefinitionResponse struct {
	// ID identifies the furniture definition.
	ID int64 `json:"id" required:"true"`
	// SpriteID identifies the Nitro rendering class.
	SpriteID int `json:"spriteId" required:"true"`
	// Name stores the technical furniture name.
	Name string `json:"name" required:"true"`
	// PublicName stores the display name.
	PublicName string `json:"publicName" required:"true"`
}

// CatalogDefinitionsResponse contains definitions missing active offers.
type CatalogDefinitionsResponse []CatalogDefinitionResponse

// CatalogRefreshResponse contains catalog publication counts.
type CatalogRefreshResponse struct {
	// Connections stores successful deliveries.
	Connections int `json:"connections" required:"true"`
	// Failures stores failed deliveries.
	Failures int `json:"failures" required:"true"`
}
