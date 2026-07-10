package service

import (
	"sync/atomic"

	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
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
}

// catalogCache exposes atomically replaceable catalog data.
type catalogCache struct {
	// snapshot points to the current immutable generation.
	snapshot atomic.Pointer[cacheSnapshot]
}

// newCache creates an empty catalog cache.
func newCache() *catalogCache {
	cache := &catalogCache{}
	cache.replace(nil, nil)

	return cache
}

// replace builds and installs one complete catalog generation.
func (cache *catalogCache) replace(pages []catalogmodel.Page, items []catalogmodel.Item) {
	snapshot := &cacheSnapshot{
		pages: make(map[int64]catalogmodel.Page, len(pages)), pageOrder: append([]catalogmodel.Page{}, pages...),
		items: make(map[int64]catalogmodel.Item, len(items)), itemsByPage: make(map[int64][]catalogmodel.Item),
	}
	for _, page := range pages {
		snapshot.pages[page.ID] = page
	}
	for _, item := range items {
		snapshot.items[item.ID] = item
		snapshot.itemsByPage[item.PageID] = append(snapshot.itemsByPage[item.PageID], item)
	}

	cache.snapshot.Store(snapshot)
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
