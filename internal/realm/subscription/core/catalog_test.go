package core

import (
	"context"
	"time"

	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
)

// fakeCatalog supplies focused catalog reward behavior.
type fakeCatalog struct {
	Catalog
	// pages stores visible catalog pages.
	pages []catalogmodel.Page
	// items stores visible page items.
	items map[int64][]catalogmodel.Item
}

// Purchase returns an empty free reward.
func (*fakeCatalog) Purchase(context.Context, catalogservice.PurchaseParams) (catalogservice.PurchaseResult, error) {
	return catalogservice.PurchaseResult{}, nil
}

// CreditsSpentSince returns deterministic eligible spending.
func (*fakeCatalog) CreditsSpentSince(context.Context, int64, time.Time) (int64, error) {
	return 100, nil
}

// CreditsSpentBetween returns deterministic period spending.
func (*fakeCatalog) CreditsSpentBetween(context.Context, int64, time.Time, time.Time) (int64, error) {
	return 100, nil
}

// Pages returns configured visible pages.
func (catalog *fakeCatalog) Pages(context.Context, int64, bool) ([]catalogmodel.Page, error) {
	return catalog.pages, nil
}

// Page returns one configured page and its items.
func (catalog *fakeCatalog) Page(_ context.Context, pageID int64, _ int64, _ bool) (catalogmodel.Page, []catalogmodel.Item, error) {
	for _, page := range catalog.pages {
		if page.ID == pageID {
			return page, catalog.items[pageID], nil
		}
	}
	return catalogmodel.Page{}, nil, catalogservice.ErrPageNotFound
}

// fakeFurniture supplies unused furniture behavior.
type fakeFurniture struct {
	furnitureservice.DefinitionGranter
}

// FindDefinitionByID returns no definition.
func (fakeFurniture) FindDefinitionByID(context.Context, int64) (furnituremodel.Definition, bool, error) {
	return furnituremodel.Definition{}, false, nil
}

// ListDefinitions returns no definitions.
func (fakeFurniture) ListDefinitions(context.Context) ([]furnituremodel.Definition, error) {
	return nil, nil
}

// Grant returns no furniture.
func (fakeFurniture) Grant(context.Context, furnitureservice.GrantParams) ([]furnituremodel.Item, error) {
	return nil, nil
}
