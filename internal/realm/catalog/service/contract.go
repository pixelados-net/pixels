// Package service contains catalog browsing and purchase behavior.
package service

import (
	"context"

	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
)

// Reader reads player-visible catalog data.
type Reader interface {
	// Pages returns pages visible to one player capability set.
	Pages(ctx context.Context, playerID int64, hasClub bool) ([]catalogmodel.Page, error)

	// Page returns one visible page and its enabled offers.
	Page(ctx context.Context, pageID int64, playerID int64, hasClub bool) (catalogmodel.Page, []catalogmodel.Item, error)

	// Definition returns cached furniture metadata for one catalog offer.
	Definition(ctx context.Context, definitionID int64) (furnituremodel.Definition, bool, error)

	// SanitizeList returns definitions without an enabled active offer.
	SanitizeList(ctx context.Context) ([]furnituremodel.Definition, error)
}

// Manager reads catalog data and processes purchases.
type Manager interface {
	Reader

	// Purchase buys one catalog offer.
	Purchase(ctx context.Context, params PurchaseParams) (PurchaseResult, error)

	// Refresh reloads the complete catalog cache.
	Refresh(ctx context.Context) error
}

// PurchaseParams contains one catalog purchase request.
type PurchaseParams struct {
	// PlayerID identifies the buyer.
	PlayerID int64

	// CatalogItemID identifies the requested offer.
	CatalogItemID int64

	// HasClub reports whether the buyer has active club membership.
	HasClub bool
}

// PurchaseResult contains one completed purchase.
type PurchaseResult struct {
	// Item stores the purchased offer snapshot.
	Item catalogmodel.Item

	// GrantedItems stores created furniture instances.
	GrantedItems []furnituremodel.Item

	// LimitedUnitNumber stores the optional LTD edition number.
	LimitedUnitNumber *int32

	// NewCreditsBalance stores the resulting credits balance.
	NewCreditsBalance int64

	// NewPointsBalance stores the resulting activity-points balance.
	NewPointsBalance int64
}

// managerAssertion verifies Service implements Manager.
var managerAssertion Manager = (*Service)(nil)
