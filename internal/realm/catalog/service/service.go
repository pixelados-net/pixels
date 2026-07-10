package service

import (
	"context"
	"fmt"

	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	catalogrepo "github.com/niflaot/pixels/internal/realm/catalog/repository"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	currencyservice "github.com/niflaot/pixels/internal/realm/inventory/currency/service"
	"github.com/niflaot/pixels/pkg/bus"
	"go.uber.org/zap"
)

// Service implements catalog browsing and purchase behavior.
type Service struct {
	// store persists catalog records and transaction boundaries.
	store catalogrepo.Store

	// currencies charges player balances.
	currencies currencyservice.Granter

	// furniture grants purchased inventory instances.
	furniture furnitureservice.Granter

	// cache stores one immutable catalog generation.
	cache *catalogCache

	// events publishes completed purchase facts.
	events bus.Publisher

	// log records post-commit projection failures.
	log *zap.Logger
}

// New creates a catalog service.
func New(store catalogrepo.Store, currencies currencyservice.Granter, furniture furnitureservice.Granter, events bus.Publisher, log *zap.Logger) *Service {
	if log == nil {
		log = zap.NewNop()
	}

	return &Service{store: store, currencies: currencies, furniture: furniture, cache: newCache(), events: events, log: log}
}

// Refresh reloads the complete catalog cache.
func (service *Service) Refresh(ctx context.Context) error {
	pages, err := service.store.ListPages(ctx)
	if err != nil {
		return fmt.Errorf("refresh catalog pages: %w", err)
	}
	items, err := service.store.ListItems(ctx, nil)
	if err != nil {
		return fmt.Errorf("refresh catalog items: %w", err)
	}

	service.cache.replace(pages, items)

	return nil
}

// Pages returns pages visible to one player capability set.
func (service *Service) Pages(_ context.Context, rank int32, hasClub bool) ([]catalogmodel.Page, error) {
	pages := service.cache.pages()
	visible := make([]catalogmodel.Page, 0, len(pages))
	for _, page := range pages {
		if service.pageAccessible(page, rank, hasClub) {
			visible = append(visible, page)
		}
	}

	return visible, nil
}

// Page returns one visible page and its enabled offers.
func (service *Service) Page(_ context.Context, pageID int64, rank int32, hasClub bool) (catalogmodel.Page, []catalogmodel.Item, error) {
	page, found := service.cache.page(pageID)
	if !found {
		return catalogmodel.Page{}, nil, ErrPageNotFound
	}
	if !service.pageAccessible(page, rank, hasClub) {
		return catalogmodel.Page{}, nil, ErrOfferNotVisible
	}

	items := service.cache.pageItems(pageID)
	visible := make([]catalogmodel.Item, 0, len(items))
	for _, item := range items {
		if item.Enabled && (!item.ClubOnly || hasClub) {
			visible = append(visible, item)
		}
	}

	return page, visible, nil
}

// SanitizeList returns definitions without an enabled active offer.
func (service *Service) SanitizeList(ctx context.Context) ([]furnituremodel.Definition, error) {
	return service.store.SanitizeList(ctx)
}

// pageAccessible verifies a page and every cached ancestor.
func (service *Service) pageAccessible(page catalogmodel.Page, rank int32, hasClub bool) bool {
	visited := make(map[int64]struct{})
	for {
		if _, found := visited[page.ID]; found {
			return false
		}
		visited[page.ID] = struct{}{}
		if !page.Accessible(rank, hasClub) {
			return false
		}
		if page.ParentID == nil {
			return true
		}
		parent, found := service.cache.page(*page.ParentID)
		if !found {
			return false
		}
		page = parent
	}
}
