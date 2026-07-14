package routes

import (
	"time"

	"github.com/niflaot/pixels/internal/permission"
	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
)

// PageResponse contains one catalog page administration record.
type PageResponse struct {
	// ID identifies the page.
	ID int64 `json:"id"`
	// ParentID identifies the optional parent page.
	ParentID *int64 `json:"parentId"`
	// Name stores the localization slug.
	Name string `json:"name"`
	// Layout identifies the client layout.
	Layout string `json:"layout"`
	// IconColor stores the client icon color.
	IconColor int32 `json:"iconColor"`
	// IconImage stores the client icon image.
	IconImage int32 `json:"iconImage"`
	// RequiredNode stores the optional page access permission.
	RequiredNode *permission.Node `json:"requiredNode,omitempty"`
	// OrderNum stores sibling display order.
	OrderNum int32 `json:"orderNum"`
	// Visible reports page tree visibility.
	Visible bool `json:"visible"`
	// Enabled reports page availability.
	Enabled bool `json:"enabled"`
	// ClubOnly reports club access policy.
	ClubOnly bool `json:"clubOnly"`
	// Version stores the optimistic locking version.
	Version int64 `json:"version"`
	// UpdatedAt stores the last mutation time.
	UpdatedAt time.Time `json:"updatedAt"`
}

// ItemResponse contains one catalog offer administration record.
type ItemResponse struct {
	// ID identifies the offer.
	ID int64 `json:"id"`
	// PageID identifies the owning page.
	PageID int64 `json:"pageId"`
	// DefinitionID identifies the furniture definition.
	DefinitionID int64 `json:"definitionId"`
	// RoomBundleTemplateRoomID identifies the cloned room template.
	RoomBundleTemplateRoomID *int64 `json:"roomBundleTemplateRoomId,omitempty"`
	// GrantsEffectID identifies the effect reward.
	GrantsEffectID *int32 `json:"grantsEffectId,omitempty"`
	// GrantsEffectDurationSeconds stores one charge duration.
	GrantsEffectDurationSeconds int32 `json:"grantsEffectDurationSeconds"`
	// Name stores the localization slug.
	Name string `json:"name"`
	// CostCredits stores the credits price.
	CostCredits int64 `json:"costCredits"`
	// CostPoints stores the activity-points price.
	CostPoints int64 `json:"costPoints"`
	// PointsType identifies the activity-points currency.
	PointsType int32 `json:"pointsType"`
	// Amount stores the furniture quantity.
	Amount int32 `json:"amount"`
	// LimitedStack stores total numbered stock.
	LimitedStack int32 `json:"limitedStack"`
	// LimitedSells stores committed numbered sales.
	LimitedSells int32 `json:"limitedSells"`
	// BundleDiscountEnabled reports bulk discount eligibility.
	BundleDiscountEnabled bool `json:"bundleDiscountEnabled"`
	// Giftable reports gift eligibility.
	Giftable bool `json:"giftable"`
	// ClubOnly reports club access policy.
	ClubOnly bool `json:"clubOnly"`
	// OrderNum stores page display order.
	OrderNum int32 `json:"orderNum"`
	// Enabled reports offer availability.
	Enabled bool `json:"enabled"`
	// ExtraData stores initial furniture state.
	ExtraData string `json:"extraData"`
	// Version stores the optimistic locking version.
	Version int64 `json:"version"`
}

// DefinitionResponse contains a furniture definition missing an active offer.
type DefinitionResponse struct {
	// ID identifies the furniture definition.
	ID int64 `json:"id"`
	// SpriteID identifies the Nitro rendering class.
	SpriteID int `json:"spriteId"`
	// Name stores the technical furniture name.
	Name string `json:"name"`
	// PublicName stores the display name.
	PublicName string `json:"publicName"`
}

// RefreshResponse contains catalog publication counts.
type RefreshResponse struct {
	// Connections stores successful publication deliveries.
	Connections int `json:"connections"`
	// Failures stores failed publication deliveries.
	Failures int `json:"failures"`
}

// pageResponse maps one catalog page record.
func pageResponse(page catalogmodel.Page) PageResponse {
	return PageResponse{ID: page.ID, ParentID: page.ParentID, Name: page.Name, Layout: page.Layout, IconColor: page.IconColor,
		IconImage: page.IconImage, RequiredNode: page.RequiredNode, OrderNum: page.OrderNum, Visible: page.Visible,
		Enabled: page.Enabled, ClubOnly: page.ClubOnly, Version: page.Version.Version, UpdatedAt: page.UpdatedAt}
}

// itemResponse maps one catalog offer record.
func itemResponse(item catalogmodel.Item) ItemResponse {
	return ItemResponse{ID: item.ID, PageID: item.PageID, DefinitionID: item.DefinitionID, RoomBundleTemplateRoomID: item.RoomBundleTemplateRoomID, GrantsEffectID: item.GrantsEffectID, GrantsEffectDurationSeconds: item.GrantsEffectDurationSeconds, Name: item.Name,
		CostCredits: item.CostCredits, CostPoints: item.CostPoints, PointsType: item.PointsType, Amount: item.Amount,
		LimitedStack: item.LimitedStack, LimitedSells: item.LimitedSells, BundleDiscountEnabled: item.BundleDiscountEnabled, Giftable: item.Giftable, ClubOnly: item.ClubOnly,
		OrderNum: item.OrderNum, Enabled: item.Enabled, ExtraData: item.ExtraData, Version: item.Version.Version}
}

// definitionResponse maps one furniture definition record.
func definitionResponse(definition furnituremodel.Definition) DefinitionResponse {
	return DefinitionResponse{ID: definition.ID, SpriteID: definition.SpriteID, Name: definition.Name, PublicName: definition.PublicName}
}
