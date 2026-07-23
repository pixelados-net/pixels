package bundle

import (
	"context"
	"errors"
	"testing"

	furnitureservice "github.com/niflaot/pixels/internal/realm/furniture/service"
	roommodel "github.com/niflaot/pixels/internal/realm/room/record/model"
	"github.com/niflaot/pixels/internal/realm/room/world/layout"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// fakeStore records bundle persistence behavior.
type fakeStore struct {
	// owned stores the buyer room count.
	owned int
	// cloned reports whether cloning ran.
	cloned bool
	// recorded stores bundle provenance.
	recorded PurchaseRecord
	// references stores active catalog links to the template.
	references int
	// template stores the room returned by template administration.
	template roommodel.Room
	// templateFound reports whether template administration finds a room.
	templateFound bool
	// templates stores the template listing.
	templates []roommodel.Room
}

// WithinTransaction runs work synchronously.
func (*fakeStore) WithinTransaction(ctx context.Context, work func(context.Context) error) error {
	return work(ctx)
}

// LockRoomOwner completes the in-memory lock.
func (*fakeStore) LockRoomOwner(context.Context, int64) error { return nil }

// CountRoomsByOwner returns the configured room count.
func (store *fakeStore) CountRoomsByOwner(context.Context, int64) (int, error) {
	return store.owned, nil
}

// CloneBundleRoom creates one test room.
func (store *fakeStore) CloneBundleRoom(context.Context, int64, int64, string) (roommodel.Room, error) {
	store.cloned = true
	return roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 44}}, Name: "Starter Loft"}, nil
}

// RecordBundlePurchase stores test provenance.
func (store *fakeStore) RecordBundlePurchase(_ context.Context, record PurchaseRecord) error {
	store.recorded = record
	return nil
}

// SetBundleTemplate returns a test room.
func (store *fakeStore) SetBundleTemplate(context.Context, int64, bool) (roommodel.Room, bool, error) {
	return store.template, store.templateFound, nil
}

// CountActiveBundleReferences returns configured active references.
func (store *fakeStore) CountActiveBundleReferences(context.Context, int64) (int, error) {
	return store.references, nil
}

// ListBundleTemplateRooms returns configured templates.
func (store *fakeStore) ListBundleTemplateRooms(context.Context) ([]roommodel.Room, error) {
	return store.templates, nil
}

// fakeRooms resolves a marked template.
type fakeRooms struct {
	// room overrides the default marked template.
	room *roommodel.Room
}

// FindByID returns a marked template.
func (rooms fakeRooms) FindByID(context.Context, int64) (roommodel.Room, bool, error) {
	if rooms.room != nil {
		return *rooms.room, true, nil
	}
	return roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 10}}, IsBundleTemplate: true}, true, nil
}

// ListByOwner returns no rooms.
func (fakeRooms) ListByOwner(context.Context, int64) ([]roommodel.Room, error) { return nil, nil }

// ListPopular returns no rooms.
func (fakeRooms) ListPopular(context.Context, int) ([]roommodel.Room, error) { return nil, nil }

// ListHighestScore returns no rooms.
func (fakeRooms) ListHighestScore(context.Context, int) ([]roommodel.Room, error) { return nil, nil }

// Search returns no rooms.
func (fakeRooms) Search(context.Context, string, int) ([]roommodel.Room, error) { return nil, nil }

// ListTags returns no room tags.
func (fakeRooms) ListTags(context.Context, int64) ([]roommodel.Tag, error) { return nil, nil }

// fakeLayouts records custom layout cloning.
type fakeLayouts struct {
	// saved stores the clone input.
	saved layout.CustomSaveParams
}

// FindCustomByRoomID returns one custom layout.
func (*fakeLayouts) FindCustomByRoomID(context.Context, int64) (layout.Layout, bool, error) {
	return layout.Layout{Heightmap: "00\r00", DoorDirection: 2, WallHeight: -1}, true, nil
}

// SaveCustom records custom geometry.
func (layouts *fakeLayouts) SaveCustom(_ context.Context, params layout.CustomSaveParams) (layout.Layout, error) {
	layouts.saved = params
	return layout.Layout{}, nil
}

// fakeFurniture clones and previews fixed products.
type fakeFurniture struct{}

// CloneRoom returns a fixed count.
func (fakeFurniture) CloneRoom(context.Context, int64, int64, int64) (int, error) { return 7, nil }

// PreviewRoom returns one grouped product.
func (fakeFurniture) PreviewRoom(context.Context, int64) ([]furnitureservice.RoomBundleProduct, error) {
	return []furnitureservice.RoomBundleProduct{{DefinitionID: 3, Quantity: 2}}, nil
}

// fakeBots clones a fixed number of template bots.
type fakeBots struct{}

// CloneRoom returns a fixed bot count.
func (fakeBots) CloneRoom(context.Context, int64, int64, int64) (int, error) { return 3, nil }

// TestCloneCopiesCustomLayoutFurnitureAndAudit verifies complete cloning.
func TestCloneCopiesCustomLayoutFurnitureAndAudit(t *testing.T) {
	store := &fakeStore{}
	layouts := &fakeLayouts{}
	service := &Service{config: Config{CloneBots: true}, store: store, rooms: fakeRooms{}, layouts: layouts, furniture: fakeFurniture{}, bots: fakeBots{}}
	result, err := service.Clone(context.Background(), CloneParams{TemplateRoomID: 10, BuyerPlayerID: 7, BuyerName: "demo", CatalogItemID: 1101})
	if err != nil || result.Room.ID != 44 || result.FurnitureCount != 7 || result.BotCount != 3 {
		t.Fatalf("result=%#v error=%v", result, err)
	}
	if !store.cloned || layouts.saved.RoomID != 44 || store.recorded.FurnitureCount != 7 || store.recorded.BotCount != 3 {
		t.Fatalf("audit=%#v layout=%#v", store.recorded, layouts.saved)
	}
}

// TestCloneRejectsRoomLimitBeforeCreation verifies the guarded limit.
func TestCloneRejectsRoomLimitBeforeCreation(t *testing.T) {
	store := &fakeStore{owned: 100}
	service := &Service{store: store, rooms: fakeRooms{}, layouts: &fakeLayouts{}, furniture: fakeFurniture{}}
	_, err := service.Clone(context.Background(), CloneParams{TemplateRoomID: 10, BuyerPlayerID: 7, BuyerName: "demo", CatalogItemID: 1101})
	if !errors.Is(err, ErrRoomLimitReached) || store.cloned {
		t.Fatalf("cloned=%v error=%v", store.cloned, err)
	}
}

// TestUnmarkRejectsReferencedTemplate verifies active offers protect templates.
func TestUnmarkRejectsReferencedTemplate(t *testing.T) {
	store := &fakeStore{references: 1}
	service := &Service{store: store, rooms: fakeRooms{}, layouts: &fakeLayouts{}, furniture: fakeFurniture{}}
	_, err := service.Unmark(context.Background(), 10)
	if !errors.Is(err, ErrTemplateReferenced) {
		t.Fatalf("error=%v", err)
	}
}

// TestTemplateAdministration verifies marking, listing, and unmarking templates.
func TestTemplateAdministration(t *testing.T) {
	template := roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 10}}, IsBundleTemplate: true}
	store := &fakeStore{template: template, templateFound: true, templates: []roommodel.Room{template}}
	service := &Service{store: store, rooms: fakeRooms{}, layouts: &fakeLayouts{}, furniture: fakeFurniture{}}
	marked, err := service.Mark(context.Background(), 10)
	if err != nil || marked.ID != 10 {
		t.Fatalf("marked=%#v error=%v", marked, err)
	}
	templates, err := service.Templates(context.Background())
	if err != nil || len(templates) != 1 {
		t.Fatalf("templates=%#v error=%v", templates, err)
	}
	unmarked, err := service.Unmark(context.Background(), 10)
	if err != nil || unmarked.ID != 10 {
		t.Fatalf("unmarked=%#v error=%v", unmarked, err)
	}
}

// TestTemplateAdministrationRejectsMissingRooms verifies missing template guards.
func TestTemplateAdministrationRejectsMissingRooms(t *testing.T) {
	service := &Service{store: &fakeStore{}, rooms: fakeRooms{}, layouts: &fakeLayouts{}, furniture: fakeFurniture{}}
	if _, err := service.Mark(context.Background(), 99); !errors.Is(err, ErrRoomNotFound) {
		t.Fatalf("mark error=%v", err)
	}
	if _, err := service.Unmark(context.Background(), 99); !errors.Is(err, ErrRoomNotFound) {
		t.Fatalf("unmark error=%v", err)
	}
}

// TestPreviewGroupsFurnitureAndValidatesTemplate verifies preview mapping and guards.
func TestPreviewGroupsFurnitureAndValidatesTemplate(t *testing.T) {
	service := &Service{store: &fakeStore{}, rooms: fakeRooms{}, layouts: &fakeLayouts{}, furniture: fakeFurniture{}}
	products, err := service.Preview(context.Background(), 10)
	if err != nil || len(products) != 1 || products[0].DefinitionID != 3 || products[0].Quantity != 2 {
		t.Fatalf("products=%#v error=%v", products, err)
	}
	plain := roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 11}}}
	service.rooms = fakeRooms{room: &plain}
	if _, err = service.Preview(context.Background(), 11); !errors.Is(err, ErrInvalidTemplate) {
		t.Fatalf("invalid template error=%v", err)
	}
}

// TestNewCreatesService verifies constructor dependency wiring.
func TestNewCreatesService(t *testing.T) {
	if service := New(Config{}, nil, nil, nil, nil, nil); service == nil {
		t.Fatal("expected service")
	}
}

// BenchmarkPreview measures grouped preview allocations.
func BenchmarkPreview(b *testing.B) {
	service := &Service{store: &fakeStore{}, rooms: fakeRooms{}, layouts: &fakeLayouts{}, furniture: fakeFurniture{}}
	b.ReportAllocs()
	for b.Loop() {
		_, _ = service.Preview(context.Background(), 10)
	}
}
