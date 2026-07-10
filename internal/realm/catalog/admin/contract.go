// Package admin manages catalog administration behavior.
package admin

import (
	"context"

	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
)

// Manager manages persistent catalog pages and offers.
type Manager interface {
	// Pages lists all active catalog pages without player access filtering.
	Pages(ctx context.Context) ([]catalogmodel.Page, error)
	// CreatePage creates one catalog page.
	CreatePage(ctx context.Context, input PageInput) (catalogmodel.Page, error)
	// UpdatePage applies a partial catalog page update.
	UpdatePage(ctx context.Context, id int64, patch PagePatch) (catalogmodel.Page, error)
	// Items lists active offers with an optional page filter.
	Items(ctx context.Context, pageID *int64) ([]catalogmodel.Item, error)
	// CreateItem creates one catalog offer and its optional LTD stock.
	CreateItem(ctx context.Context, input ItemInput) (catalogmodel.Item, error)
	// UpdateItem applies a partial catalog offer update.
	UpdateItem(ctx context.Context, id int64, patch ItemPatch) (catalogmodel.Item, error)
	// DeleteItem soft deletes one catalog offer.
	DeleteItem(ctx context.Context, id int64) error
	// SanitizeList lists furniture definitions without active offers.
	SanitizeList(ctx context.Context) ([]furnituremodel.Definition, error)
	// Refresh reloads the player-facing catalog cache.
	Refresh(ctx context.Context) error
}

// PageInput contains writable catalog page fields.
type PageInput struct {
	// ParentID identifies the optional parent page.
	ParentID *int64
	// Name stores the stable localization slug.
	Name string
	// Layout identifies the client layout.
	Layout string
	// IconColor stores the client icon color.
	IconColor int32
	// IconImage stores the client icon image.
	IconImage int32
	// MinRank stores the minimum visible rank.
	MinRank int32
	// OrderNum stores sibling display order.
	OrderNum int32
	// Visible reports whether the page appears in the tree.
	Visible bool
	// Enabled reports whether the page can be opened.
	Enabled bool
	// ClubOnly reports whether club membership is required.
	ClubOnly bool
}

// PagePatch contains optional writable catalog page fields.
type PagePatch struct {
	// ParentID replaces the optional parent id when present.
	ParentID **int64
	// Name replaces the localization slug when present.
	Name *string
	// Layout replaces the client layout when present.
	Layout *string
	// IconColor replaces the icon color when present.
	IconColor *int32
	// IconImage replaces the icon image when present.
	IconImage *int32
	// MinRank replaces the minimum rank when present.
	MinRank *int32
	// OrderNum replaces sibling display order when present.
	OrderNum *int32
	// Visible replaces page visibility when present.
	Visible *bool
	// Enabled replaces page availability when present.
	Enabled *bool
	// ClubOnly replaces club access policy when present.
	ClubOnly *bool
}

// ItemInput contains writable catalog offer fields.
type ItemInput struct {
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
	// Amount stores the furniture quantity granted.
	Amount int32
	// LimitedStack stores numbered stock.
	LimitedStack int32
	// OfferID stores an optional future grouping id.
	OfferID *int64
	// ClubOnly reports whether club membership is required.
	ClubOnly bool
	// OrderNum stores page display order.
	OrderNum int32
	// Enabled reports whether the offer can be purchased.
	Enabled bool
	// ExtraData stores initial furniture protocol state.
	ExtraData string
}

// ItemPatch contains optional writable catalog offer fields.
type ItemPatch struct {
	// PageID replaces the owning page when present.
	PageID *int64
	// DefinitionID replaces the furniture definition when present.
	DefinitionID *int64
	// Name replaces the localization slug when present.
	Name *string
	// CostCredits replaces the credits price when present.
	CostCredits *int64
	// CostPoints replaces the points price when present.
	CostPoints *int64
	// PointsType replaces the points currency when present.
	PointsType *int32
	// Amount replaces the grant quantity when present.
	Amount *int32
	// LimitedStack replaces numbered stock when present.
	LimitedStack *int32
	// OfferID replaces the optional grouping id when present.
	OfferID **int64
	// ClubOnly replaces club access policy when present.
	ClubOnly *bool
	// OrderNum replaces page display order when present.
	OrderNum *int32
	// Enabled replaces offer availability when present.
	Enabled *bool
	// ExtraData replaces initial furniture state when present.
	ExtraData *string
}
