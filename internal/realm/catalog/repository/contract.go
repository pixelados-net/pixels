// Package repository contains PostgreSQL access for catalog records.
package repository

import (
	"context"

	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
)

// PageReader reads catalog pages.
type PageReader interface {
	// ListPages lists every active catalog page.
	ListPages(ctx context.Context) ([]catalogmodel.Page, error)

	// FindPageByID finds one active catalog page.
	FindPageByID(ctx context.Context, id int64) (catalogmodel.Page, bool, error)
}

// PageWriter writes catalog pages.
type PageWriter interface {
	// CreatePage creates one catalog page.
	CreatePage(ctx context.Context, page catalogmodel.Page) (catalogmodel.Page, error)

	// UpdatePage updates one page using optimistic locking.
	UpdatePage(ctx context.Context, page catalogmodel.Page) (catalogmodel.Page, bool, error)
}

// ItemReader reads catalog offers.
type ItemReader interface {
	// ListItems lists active offers, optionally restricted to one page.
	ListItems(ctx context.Context, pageID *int64) ([]catalogmodel.Item, error)

	// FindItemByID finds one active catalog offer.
	FindItemByID(ctx context.Context, id int64) (catalogmodel.Item, bool, error)

	// SanitizeList lists active furniture definitions without an active offer.
	SanitizeList(ctx context.Context) ([]furnituremodel.Definition, error)

	// CountEnabledDefinitionsWithoutOffer counts active definitions without enabled offers.
	CountEnabledDefinitionsWithoutOffer(ctx context.Context) (int64, error)
}

// ItemWriter writes catalog offers.
type ItemWriter interface {
	// CreateItem creates one catalog offer.
	CreateItem(ctx context.Context, item catalogmodel.Item) (catalogmodel.Item, error)

	// UpdateItem updates one offer using optimistic locking.
	UpdateItem(ctx context.Context, item catalogmodel.Item) (catalogmodel.Item, bool, error)

	// SoftDeleteItem soft deletes one offer using optimistic locking.
	SoftDeleteItem(ctx context.Context, id int64, version int64) (bool, error)
}

// LimitedWriter manages numbered LTD allocations.
type LimitedWriter interface {
	// CreateLimitedUnits creates numbered units for an LTD offer.
	CreateLimitedUnits(ctx context.Context, catalogItemID int64, quantity int32) error

	// ReserveLimitedUnit atomically reserves the lowest available LTD number.
	ReserveLimitedUnit(ctx context.Context, catalogItemID int64, playerID int64) (int32, bool, error)

	// CompleteLimitedUnit links a reservation and advances the offer sale count.
	CompleteLimitedUnit(ctx context.Context, catalogItemID int64, unitNumber int32, playerID int64, furnitureItemID int64) (bool, error)
}

// Store reads and mutates catalog persistence.
type Store interface {
	PageReader
	PageWriter
	ItemReader
	ItemWriter
	LimitedWriter

	// WithinTransaction runs catalog purchase work atomically.
	WithinTransaction(ctx context.Context, work func(context.Context) error) error
}

// storeAssertion verifies Repository implements Store.
var storeAssertion Store = (*Repository)(nil)
