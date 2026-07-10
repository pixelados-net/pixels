package service

import (
	"context"
	"fmt"

	permissionservice "github.com/niflaot/pixels/internal/permission/service"
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
	furniture furnitureservice.DefinitionGranter

	// cache stores one immutable catalog generation.
	cache *catalogCache

	// events publishes completed purchase facts.
	events bus.Publisher

	// log records post-commit projection failures.
	log *zap.Logger
	// permissions resolves optional catalog page requirements.
	permissions permissionservice.Checker
}

// New creates a catalog service.
func New(store catalogrepo.Store, currencies currencyservice.Granter, furniture furnitureservice.DefinitionGranter, events bus.Publisher, log *zap.Logger, checkers ...permissionservice.Checker) *Service {
	if log == nil {
		log = zap.NewNop()
	}

	service := &Service{store: store, currencies: currencies, furniture: furniture, cache: newCache(), events: events, log: log}
	if len(checkers) > 0 {
		service.permissions = checkers[0]
	}

	return service
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
	definitions, err := service.furniture.ListDefinitions(ctx)
	if err != nil {
		return fmt.Errorf("refresh catalog furniture definitions: %w", err)
	}

	service.cache.replace(pages, items, definitions)

	return nil
}

// Definition returns cached furniture metadata for one catalog offer.
func (service *Service) Definition(_ context.Context, definitionID int64) (furnituremodel.Definition, bool, error) {
	definition, found := service.cache.definition(definitionID)

	return definition, found, nil
}

// Pages returns pages visible to one player capability set.
func (service *Service) Pages(ctx context.Context, playerID int64, hasClub bool) ([]catalogmodel.Page, error) {
	pages := service.cache.pages()
	visible := make([]catalogmodel.Page, 0, len(pages))
	for _, page := range pages {
		accessible, err := service.pageAccessible(ctx, page, playerID, hasClub)
		if err != nil {
			return nil, err
		}
		if accessible {
			visible = append(visible, page)
		}
	}

	return visible, nil
}

// Page returns one visible page and its enabled offers.
func (service *Service) Page(ctx context.Context, pageID int64, playerID int64, hasClub bool) (catalogmodel.Page, []catalogmodel.Item, error) {
	page, found := service.cache.page(pageID)
	if !found {
		return catalogmodel.Page{}, nil, ErrPageNotFound
	}
	accessible, err := service.pageAccessible(ctx, page, playerID, hasClub)
	if err != nil {
		return catalogmodel.Page{}, nil, err
	}
	if !accessible {
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
func (service *Service) pageAccessible(ctx context.Context, page catalogmodel.Page, playerID int64, hasClub bool) (bool, error) {
	visited := make(map[int64]struct{})
	for {
		if _, found := visited[page.ID]; found {
			return false, nil
		}
		visited[page.ID] = struct{}{}
		if !page.Accessible(hasClub) {
			return false, nil
		}
		if page.RequiredNode != nil {
			if service.permissions == nil {
				return false, nil
			}
			allowed, err := service.permissions.HasPermission(ctx, playerID, *page.RequiredNode)
			if err != nil {
				return false, err
			}
			if !allowed {
				return false, nil
			}
		}
		if page.ParentID == nil {
			return true, nil
		}
		parent, found := service.cache.page(*page.ParentID)
		if !found {
			return false, nil
		}
		page = parent
	}
}
