package layout

import (
	"context"
	"errors"
	"testing"
)

// validLayoutForTest returns valid editable layout values.
func validLayoutForTest() Layout {
	return Layout{Name: "model_a", TileSize: 12, Heightmap: "xxx\rx0x\rxxx", DoorX: 1, DoorY: 1, DoorDirection: 2, Enabled: true}
}

// TestServiceCreateNormalizesLayout verifies editable layout creation.
func TestServiceCreateNormalizesLayout(t *testing.T) {
	store := newFakeStore()

	roomLayout, err := NewService(store).Create(context.Background(), SaveParams{
		Name:          "a",
		TileSize:      12,
		Heightmap:     " xxx\rx0x\rxxx ",
		DoorX:         1,
		DoorY:         1,
		DoorDirection: 2,
		Enabled:       true,
	})
	if err != nil {
		t.Fatalf("create layout: %v", err)
	}

	if roomLayout.Name != "model_a" || roomLayout.Heightmap != "xxx\rx0x\rxxx" {
		t.Fatalf("unexpected layout %#v", roomLayout)
	}
}

// TestServiceCreateRejectsInvalidLayout verifies layout validation.
func TestServiceCreateRejectsInvalidLayout(t *testing.T) {
	_, err := NewService(newFakeStore()).Create(context.Background(), SaveParams{Name: "a"})
	if !errors.Is(err, ErrInvalidLayout) {
		t.Fatalf("expected invalid layout, got %v", err)
	}
}

// TestServiceUpdateReportsMissingLayout verifies missing update behavior.
func TestServiceUpdateReportsMissingLayout(t *testing.T) {
	store := newFakeStore()
	store.found = false

	_, err := NewService(store).Update(context.Background(), 7, validSaveParamsForTest())
	if !errors.Is(err, ErrLayoutNotFound) {
		t.Fatalf("expected missing layout, got %v", err)
	}
}

// TestServiceFindByIDRejectsInvalidID verifies id validation.
func TestServiceFindByIDRejectsInvalidID(t *testing.T) {
	_, _, err := NewService(newFakeStore()).FindByID(context.Background(), 0)
	if !errors.Is(err, ErrInvalidLayoutID) {
		t.Fatalf("expected invalid id, got %v", err)
	}
}

// TestServiceCatalogLoadsPersistentLayouts verifies persistent catalog loading.
func TestServiceCatalogLoadsPersistentLayouts(t *testing.T) {
	catalog, err := NewService(newFakeStore()).Catalog(context.Background())
	if err != nil {
		t.Fatalf("load catalog: %v", err)
	}

	roomLayout, found := catalog.Find("a")
	if !found {
		t.Fatal("expected layout")
	}

	if roomLayout.ID != 7 {
		t.Fatalf("expected persistent id 7, got %d", roomLayout.ID)
	}
}

// TestLayoutGridParsesHeightmap verifies layout to grid conversion.
func TestLayoutGridParsesHeightmap(t *testing.T) {
	roomGrid, err := validLayoutForTest().Grid()
	if err != nil {
		t.Fatalf("parse layout grid: %v", err)
	}

	if roomGrid.Width() != 3 || roomGrid.Height() != 3 || roomGrid.ValidCount() != 1 {
		t.Fatalf("unexpected layout grid dimensions or count")
	}

	door, ok := roomGrid.Door()
	if !ok || door.X != 1 || door.Y != 1 {
		t.Fatalf("unexpected layout door %#v found=%v", door, ok)
	}
}

// newFakeStore creates a room layout store for tests.
func newFakeStore() *fakeStore {
	roomLayout := validLayoutForTest()
	roomLayout.ID = 7

	return &fakeStore{layout: roomLayout, found: true}
}

// fakeStore records room layout store calls for tests.
type fakeStore struct {
	// layout is the returned room layout.
	layout Layout

	// found reports whether lookups and updates succeed.
	found bool

	// err is returned by store calls.
	err error
}

// Create creates a room layout record for tests.
func (store *fakeStore) Create(_ context.Context, params CreateRecordParams) (Layout, error) {
	store.layout = params.Layout
	store.layout.ID = 7

	return store.layout, store.err
}

// Update updates a room layout record for tests.
func (store *fakeStore) Update(_ context.Context, params UpdateRecordParams) (Layout, bool, error) {
	store.layout = params.Layout
	store.layout.ID = params.ID

	return store.layout, store.found, store.err
}

// FindByID finds a room layout by id for tests.
func (store *fakeStore) FindByID(context.Context, int64) (Layout, bool, error) {
	return store.layout, store.found, store.err
}

// FindByName finds a room layout by name for tests.
func (store *fakeStore) FindByName(context.Context, string) (Layout, bool, error) {
	return store.layout, store.found, store.err
}

// List lists room layouts for tests.
func (store *fakeStore) List(context.Context) ([]Layout, error) {
	return []Layout{store.layout}, store.err
}

// validSaveParamsForTest returns valid editable layout input.
func validSaveParamsForTest() SaveParams {
	return SaveParams{Name: "a", TileSize: 12, Heightmap: "xxx\rx0x\rxxx", DoorX: 1, DoorY: 1, DoorDirection: 2, Enabled: true}
}
