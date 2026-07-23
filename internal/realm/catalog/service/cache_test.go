package service

import (
	"context"
	"testing"

	"github.com/niflaot/pixels/internal/permission"
	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// TestCacheReplacesCompleteGeneration verifies atomic catalog snapshots.
func TestCacheReplacesCompleteGeneration(t *testing.T) {
	cache := newCache()
	cache.replace(
		[]catalogmodel.Page{{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 1}}, Name: "chairs"}},
		[]catalogmodel.Item{{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 2}}, PageID: 1, Name: "chair_plasto"}},
		[]furnituremodel.Definition{{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 3}}, Name: "chair_plasto"}},
	)

	page, pageFound := cache.page(1)
	item, itemFound := cache.item(2)
	definition, definitionFound := cache.definition(3)
	if !pageFound || !itemFound || !definitionFound || page.Name != "chairs" || item.Name != "chair_plasto" || definition.Name != "chair_plasto" {
		t.Fatalf("unexpected page=%#v item=%#v", page, item)
	}
	if len(cache.pages()) != 1 || len(cache.pageItems(1)) != 1 {
		t.Fatal("expected one complete cache generation")
	}

	cache.replace(nil, nil, nil)
	if _, found := cache.item(2); found || len(cache.pages()) != 0 {
		t.Fatal("expected old generation to be unreachable")
	}
}

// BenchmarkCachePageItems measures immutable offer snapshot reads.
func BenchmarkCachePageItems(b *testing.B) {
	cache := newCache()
	items := make([]catalogmodel.Item, 100)
	for index := range items {
		items[index] = catalogmodel.Item{
			Base:   sharedmodel.Base{Identity: sharedmodel.Identity{ID: int64(index + 1)}},
			PageID: 1,
		}
	}
	cache.replace([]catalogmodel.Page{{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 1}}}}, items, nil)

	b.ReportAllocs()
	for b.Loop() {
		if len(cache.pageItems(1)) != 100 {
			b.Fatal("unexpected cached item count")
		}
	}
}

// BenchmarkCatalogPagePermission measures one cached page policy check.
func BenchmarkCatalogPagePermission(b *testing.B) {
	node := permission.Node("catalog.benchmark.access")
	cache := newCache()
	cache.replace([]catalogmodel.Page{{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 1}},
		Name: "staff", Visible: true, Enabled: true, RequiredNode: &node}}, nil, nil)
	service := &Service{cache: cache, permissions: fixedChecker(true)}
	ctx := context.Background()
	b.ReportAllocs()
	for b.Loop() {
		page, found := cache.page(1)
		allowed, err := service.pageAccessible(ctx, page, 7, false)
		if err != nil || !found || !allowed {
			b.Fatalf("unexpected allowed=%v found=%v err=%v", allowed, found, err)
		}
	}
}
