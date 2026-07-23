package service

import (
	"context"
	"sync/atomic"
	"time"

	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
)

// cacheSnapshot contains one immutable catalog generation.
type cacheSnapshot struct {
	// pages stores pages by durable id.
	pages map[int64]catalogmodel.Page

	// pageOrder stores pages in persistence order.
	pageOrder []catalogmodel.Page

	// items stores offers by durable id.
	items map[int64]catalogmodel.Item

	// itemsByPage stores ordered offers by page.
	itemsByPage map[int64][]catalogmodel.Item

	// definitions stores furniture metadata by durable id.
	definitions map[int64]furnituremodel.Definition

	// products stores ordered bundle products by offer id.
	products map[int64][]catalogmodel.Product
}

// catalogCache exposes atomically replaceable catalog data.
type catalogCache struct {
	// snapshot points to the current immutable generation.
	snapshot atomic.Pointer[cacheSnapshot]
}

// newCache creates an empty catalog cache.
func newCache() *catalogCache {
	cache := &catalogCache{}
	cache.replace(nil, nil, nil, nil)

	return cache
}

// replace builds and installs one complete catalog generation.
func (cache *catalogCache) replace(pages []catalogmodel.Page, items []catalogmodel.Item, definitions []furnituremodel.Definition, productSets ...[]catalogmodel.Product) {
	var products []catalogmodel.Product
	if len(productSets) != 0 {
		products = productSets[0]
	}
	snapshot := &cacheSnapshot{
		pages: make(map[int64]catalogmodel.Page, len(pages)), pageOrder: append([]catalogmodel.Page{}, pages...),
		items: make(map[int64]catalogmodel.Item, len(items)), itemsByPage: make(map[int64][]catalogmodel.Item),
		definitions: make(map[int64]furnituremodel.Definition, len(definitions)), products: make(map[int64][]catalogmodel.Product),
	}
	for _, page := range pages {
		snapshot.pages[page.ID] = page
	}
	for _, item := range items {
		snapshot.items[item.ID] = item
		snapshot.itemsByPage[item.PageID] = append(snapshot.itemsByPage[item.PageID], item)
	}
	for _, definition := range definitions {
		snapshot.definitions[definition.ID] = definition
	}
	for _, product := range products {
		snapshot.products[product.CatalogItemID] = append(snapshot.products[product.CatalogItemID], product)
	}

	cache.snapshot.Store(snapshot)
}

// products returns one offer's immutable product list.
func (cache *catalogCache) products(id int64) []catalogmodel.Product {
	return cache.snapshot.Load().products[id]
}

// definition returns one cached furniture definition.
func (cache *catalogCache) definition(id int64) (furnituremodel.Definition, bool) {
	definition, found := cache.snapshot.Load().definitions[id]

	return definition, found
}

// pages returns all cached pages in persistence order.
func (cache *catalogCache) pages() []catalogmodel.Page {
	return cache.snapshot.Load().pageOrder
}

// page returns one cached page.
func (cache *catalogCache) page(id int64) (catalogmodel.Page, bool) {
	page, found := cache.snapshot.Load().pages[id]

	return page, found
}

// item returns one cached offer.
func (cache *catalogCache) item(id int64) (catalogmodel.Item, bool) {
	item, found := cache.snapshot.Load().items[id]

	return item, found
}

// pageItems returns one page's cached offers in persistence order.
func (cache *catalogCache) pageItems(pageID int64) []catalogmodel.Item {
	return cache.snapshot.Load().itemsByPage[pageID]
}

// Item returns one cached catalog item.
func (service *Service) Item(id int64) (catalogmodel.Item, bool) { return service.cache.item(id) }

// PageExpiration returns one cached page expiration.
func (service *Service) PageExpiration(id int64, now time.Time) (catalogmodel.Page, int32, bool) {
	page, found := service.cache.page(id)
	if !found || page.ExpiresAt == nil {
		return catalogmodel.Page{}, 0, false
	}
	return page, secondsUntil(*page.ExpiresAt, now), true
}

// EarliestExpiration returns the nearest expiration among visible pages.
func (service *Service) EarliestExpiration(pages []catalogmodel.Page, now time.Time) (catalogmodel.Page, int32, bool) {
	var selected catalogmodel.Page
	found := false
	for _, page := range pages {
		if page.ExpiresAt == nil || found && !page.ExpiresAt.Before(*selected.ExpiresAt) {
			continue
		}
		selected, found = page, true
	}
	if !found {
		return catalogmodel.Page{}, 0, false
	}
	return selected, secondsUntil(*selected.ExpiresAt, now), true
}

// NextLimited returns the nearest scheduled future LTD offer.
func (service *Service) NextLimited(now time.Time) (catalogmodel.Item, int32, bool) {
	var selected catalogmodel.Item
	found := false
	for _, item := range service.cache.snapshot.Load().items {
		if !item.IsLimited() || item.ScheduledAt == nil || !item.ScheduledAt.After(now) || found && !item.ScheduledAt.Before(*selected.ScheduledAt) {
			continue
		}
		selected, found = item, true
	}
	if !found {
		return catalogmodel.Item{}, 0, false
	}
	return selected, secondsUntil(*selected.ScheduledAt, now), true
}

// secondsUntil clamps one protocol countdown to a non-negative int32.
func secondsUntil(deadline time.Time, now time.Time) int32 {
	seconds := int64(deadline.Sub(now) / time.Second)
	if seconds <= 0 {
		return 0
	}
	if seconds > int64(^uint32(0)>>1) {
		return int32(^uint32(0) >> 1)
	}
	return int32(seconds)
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
			if err != nil || !allowed {
				return false, err
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
