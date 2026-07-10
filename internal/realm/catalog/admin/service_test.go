package admin

import (
	"context"
	"errors"
	"testing"

	catalogmodel "github.com/niflaot/pixels/internal/realm/catalog/model"
	catalogrepo "github.com/niflaot/pixels/internal/realm/catalog/repository"
	catalogservice "github.com/niflaot/pixels/internal/realm/catalog/service"
	furnituremodel "github.com/niflaot/pixels/internal/realm/furniture/model"
	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// TestPageLifecycle verifies catalog page creation and partial updates.
func TestPageLifecycle(t *testing.T) {
	fixture := newFixture()
	created, err := fixture.service.CreatePage(context.Background(), PageInput{Name: "chairs", Layout: "default_3x3", MinRank: 1, Visible: true, Enabled: true})
	if err != nil || created.ID != 2 {
		t.Fatalf("unexpected created page %#v error %v", created, err)
	}
	layout := "spaces"
	updated, err := fixture.service.UpdatePage(context.Background(), created.ID, PagePatch{Layout: &layout})
	if err != nil || updated.Layout != layout || fixture.catalog.refreshes != 2 {
		t.Fatalf("unexpected updated page %#v refreshes=%d error %v", updated, fixture.catalog.refreshes, err)
	}
	pages, err := fixture.service.Pages(context.Background())
	if err != nil || len(pages) != 2 {
		t.Fatalf("unexpected pages %#v error %v", pages, err)
	}
}

// TestCreatePageRejectsMissingParent verifies parent validation.
func TestCreatePageRejectsMissingParent(t *testing.T) {
	fixture := newFixture()
	parentID := int64(99)
	_, err := fixture.service.CreatePage(context.Background(), PageInput{ParentID: &parentID, Name: "bad", Layout: "default_3x3", MinRank: 1})
	if !errors.Is(err, ErrPageNotFound) {
		t.Fatalf("expected page not found, got %v", err)
	}
}

// TestItemLifecycleVerifiesLimitedStockAndDeletion verifies offer administration.
func TestItemLifecycleVerifiesLimitedStockAndDeletion(t *testing.T) {
	fixture := newFixture()
	created, err := fixture.service.CreateItem(context.Background(), ItemInput{PageID: 1, DefinitionID: 2, Name: "chair",
		CostCredits: 2, PointsType: catalogmodel.CreditsType, Amount: 1, LimitedStack: 10, Enabled: true})
	if err != nil || created.ID != 1 || fixture.store.synced != 10 {
		t.Fatalf("unexpected created item %#v synced=%d error %v", created, fixture.store.synced, err)
	}
	price := int64(4)
	stack := int32(12)
	updated, err := fixture.service.UpdateItem(context.Background(), created.ID, ItemPatch{CostCredits: &price, LimitedStack: &stack})
	if err != nil || updated.CostCredits != 4 || fixture.store.synced != 12 {
		t.Fatalf("unexpected updated item %#v synced=%d error %v", updated, fixture.store.synced, err)
	}
	if err := fixture.service.DeleteItem(context.Background(), created.ID); err != nil {
		t.Fatalf("delete item: %v", err)
	}
	items, _ := fixture.service.Items(context.Background(), nil)
	if len(items) != 0 {
		t.Fatalf("expected deleted offer to disappear, got %#v", items)
	}
}

// TestCreateItemRejectsMissingDefinition verifies furniture reference validation.
func TestCreateItemRejectsMissingDefinition(t *testing.T) {
	fixture := newFixture()
	_, err := fixture.service.CreateItem(context.Background(), ItemInput{PageID: 1, DefinitionID: 99, Name: "bad", PointsType: -1, Amount: 1})
	if !errors.Is(err, ErrDefinitionNotFound) {
		t.Fatalf("expected definition not found, got %v", err)
	}
}

// adminFixture contains administration test collaborators.
type adminFixture struct {
	// service stores tested behavior.
	service *Service
	// store stores fake persistence.
	store *adminStore
	// catalog stores cache refresh calls.
	catalog *adminCatalog
}

// newFixture creates catalog administration fixtures.
func newFixture() adminFixture {
	store := &adminStore{pages: []catalogmodel.Page{{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 1}}, Name: "root", Layout: "default_3x3", MinRank: 1}}}
	catalog := &adminCatalog{}
	definitions := adminDefinitions{definitions: []furnituremodel.Definition{{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 2}}}}}

	return adminFixture{service: New(store, catalog, definitions), store: store, catalog: catalog}
}

// adminStore stores mutable administration fixtures.
type adminStore struct {
	// Store supplies unused persistence methods.
	catalogrepo.Store
	// pages stores page fixtures.
	pages []catalogmodel.Page
	// items stores offer fixtures.
	items []catalogmodel.Item
	// synced stores the latest LTD quantity.
	synced int32
}

// ListPages lists page fixtures.
func (store *adminStore) ListPages(context.Context) ([]catalogmodel.Page, error) {
	return append([]catalogmodel.Page{}, store.pages...), nil
}

// FindPageByID finds a page fixture.
func (store *adminStore) FindPageByID(_ context.Context, id int64) (catalogmodel.Page, bool, error) {
	for _, page := range store.pages {
		if page.ID == id {
			return page, true, nil
		}
	}
	return catalogmodel.Page{}, false, nil
}

// CreatePage creates a page fixture.
func (store *adminStore) CreatePage(_ context.Context, page catalogmodel.Page) (catalogmodel.Page, error) {
	page.ID = int64(len(store.pages) + 1)
	store.pages = append(store.pages, page)
	return page, nil
}

// UpdatePage updates a page fixture.
func (store *adminStore) UpdatePage(_ context.Context, page catalogmodel.Page) (catalogmodel.Page, bool, error) {
	for index := range store.pages {
		if store.pages[index].ID == page.ID {
			page.Version.Version++
			store.pages[index] = page
			return page, true, nil
		}
	}
	return catalogmodel.Page{}, false, nil
}

// ListItems lists offer fixtures.
func (store *adminStore) ListItems(context.Context, *int64) ([]catalogmodel.Item, error) {
	return append([]catalogmodel.Item{}, store.items...), nil
}

// FindItemByID finds an offer fixture.
func (store *adminStore) FindItemByID(_ context.Context, id int64) (catalogmodel.Item, bool, error) {
	for _, item := range store.items {
		if item.ID == id {
			return item, true, nil
		}
	}
	return catalogmodel.Item{}, false, nil
}

// CreateItem creates an offer fixture.
func (store *adminStore) CreateItem(_ context.Context, item catalogmodel.Item) (catalogmodel.Item, error) {
	item.ID = int64(len(store.items) + 1)
	store.items = append(store.items, item)
	return item, nil
}

// UpdateItem updates an offer fixture.
func (store *adminStore) UpdateItem(_ context.Context, item catalogmodel.Item) (catalogmodel.Item, bool, error) {
	for index := range store.items {
		if store.items[index].ID == item.ID {
			item.Version.Version++
			store.items[index] = item
			return item, true, nil
		}
	}
	return catalogmodel.Item{}, false, nil
}

// SoftDeleteItem deletes an offer fixture.
func (store *adminStore) SoftDeleteItem(_ context.Context, id int64, _ int64) (bool, error) {
	for index, item := range store.items {
		if item.ID == id {
			store.items = append(store.items[:index], store.items[index+1:]...)
			return true, nil
		}
	}
	return false, nil
}

// CreateLimitedUnits records LTD fixture stock.
func (store *adminStore) CreateLimitedUnits(_ context.Context, _ int64, quantity int32) error {
	store.synced = quantity
	return nil
}

// SyncLimitedUnits records updated LTD fixture stock.
func (store *adminStore) SyncLimitedUnits(_ context.Context, _ int64, quantity int32) error {
	store.synced = quantity
	return nil
}

// WithinTransaction executes fixture work.
func (store *adminStore) WithinTransaction(ctx context.Context, work func(context.Context) error) error {
	return work(ctx)
}

// SanitizeList returns one fixture definition.
func (store *adminStore) SanitizeList(context.Context) ([]furnituremodel.Definition, error) {
	return []furnituremodel.Definition{{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}}}}, nil
}

// adminCatalog records cache refresh calls.
type adminCatalog struct {
	catalogservice.Manager
	refreshes int
}

// Refresh records one cache refresh.
func (catalog *adminCatalog) Refresh(context.Context) error { catalog.refreshes++; return nil }

// adminDefinitions reads definition fixtures.
type adminDefinitions struct {
	furnitureservice.DefinitionGranter
	definitions []furnituremodel.Definition
}

// FindDefinitionByID finds one definition fixture.
func (definitions adminDefinitions) FindDefinitionByID(_ context.Context, id int64) (furnituremodel.Definition, bool, error) {
	for _, definition := range definitions.definitions {
		if definition.ID == id {
			return definition, true, nil
		}
	}
	return furnituremodel.Definition{}, false, nil
}
