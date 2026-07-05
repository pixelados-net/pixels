package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/niflaot/pixels/internal/realm/room/layout"
	roommodel "github.com/niflaot/pixels/internal/realm/room/model"
	"github.com/niflaot/pixels/internal/realm/room/repository"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// TestCreateValidatesLayoutAndNormalizesTags verifies room creation behavior.
func TestCreateValidatesLayoutAndNormalizesTags(t *testing.T) {
	store := newFakeStore()
	room, err := New(store, fakeLayouts{found: true, enabled: true}).Create(context.Background(), CreateParams{
		OwnerPlayerID: 7,
		OwnerName:     " demo ",
		Name:          " Test Room ",
		Description:   " hello ",
		ModelName:     "model_a",
		Tags:          []string{" Fun ", "fun", "Build"},
	})
	if err != nil {
		t.Fatalf("create room: %v", err)
	}

	if room.ID != 9 || len(store.tags) != 2 || store.tags[0] != "fun" {
		t.Fatalf("unexpected create result room=%#v tags=%#v", room, store.tags)
	}
}

// TestCreateRejectsDisabledLayout verifies disabled layouts cannot create rooms.
func TestCreateRejectsDisabledLayout(t *testing.T) {
	_, err := New(newFakeStore(), fakeLayouts{found: true}).Create(context.Background(), validCreateForTest())
	if !errors.Is(err, ErrLayoutNotAvailable) {
		t.Fatalf("expected layout error, got %v", err)
	}
}

// TestSoftDeleteReportsMissingRoom verifies delete missing behavior.
func TestSoftDeleteReportsMissingRoom(t *testing.T) {
	store := newFakeStore()
	store.deleted = false

	err := New(store, fakeLayouts{found: true, enabled: true}).SoftDelete(context.Background(), 8)
	if !errors.Is(err, ErrRoomNotFound) {
		t.Fatalf("expected room not found, got %v", err)
	}
}

// TestFindByIDRejectsInvalidID verifies id validation.
func TestFindByIDRejectsInvalidID(t *testing.T) {
	_, _, err := New(newFakeStore(), fakeLayouts{}).FindByID(context.Background(), 0)
	if !errors.Is(err, ErrInvalidRoomID) {
		t.Fatalf("expected invalid id, got %v", err)
	}
}

// TestListPopularNormalizesLimit verifies list limit normalization.
func TestListPopularNormalizesLimit(t *testing.T) {
	store := newFakeStore()
	_, err := New(store, fakeLayouts{}).ListPopular(context.Background(), 200)
	if err != nil {
		t.Fatalf("list popular: %v", err)
	}

	if store.limit != 100 {
		t.Fatalf("expected capped limit, got %d", store.limit)
	}
}

// TestListByOwnerRejectsInvalidOwner verifies owner validation.
func TestListByOwnerRejectsInvalidOwner(t *testing.T) {
	_, err := New(newFakeStore(), fakeLayouts{}).ListByOwner(context.Background(), 0)
	if !errors.Is(err, ErrInvalidOwner) {
		t.Fatalf("expected invalid owner, got %v", err)
	}
}

// TestListHighestScoreUsesDefaultLimit verifies default limit normalization.
func TestListHighestScoreUsesDefaultLimit(t *testing.T) {
	store := newFakeStore()
	_, err := New(store, fakeLayouts{}).ListHighestScore(context.Background(), 0)
	if err != nil {
		t.Fatalf("list highest score: %v", err)
	}
	if store.limit != 50 {
		t.Fatalf("expected default limit, got %d", store.limit)
	}
}

// TestListCategoriesReadsStore verifies category listing.
func TestListCategoriesReadsStore(t *testing.T) {
	categories, err := New(newFakeStore(), fakeLayouts{}).ListCategories(context.Background())
	if err != nil {
		t.Fatalf("list categories: %v", err)
	}
	if len(categories) != 1 || categories[0].Caption != "Social" {
		t.Fatalf("unexpected categories %#v", categories)
	}
}

// TestCreateRejectsInvalidInput verifies create validation.
func TestCreateRejectsInvalidInput(t *testing.T) {
	cases := []struct {
		name     string
		params   CreateParams
		expected error
	}{
		{name: "owner", params: CreateParams{Name: "Test Room", ModelName: "model_a"}, expected: ErrInvalidOwner},
		{name: "name", params: CreateParams{OwnerPlayerID: 7, OwnerName: "demo", Name: "no", ModelName: "model_a"}, expected: ErrInvalidRoomName},
		{name: "description", params: CreateParams{OwnerPlayerID: 7, OwnerName: "demo", Name: "Test Room", Description: strings.Repeat("x", MaxRoomDescriptionLength+1), ModelName: "model_a"}, expected: ErrInvalidDescription},
		{name: "max users", params: CreateParams{OwnerPlayerID: 7, OwnerName: "demo", Name: "Test Room", MaxUsers: MaxRoomUsers + 1, ModelName: "model_a"}, expected: ErrInvalidMaxUsers},
		{name: "trade", params: CreateParams{OwnerPlayerID: 7, OwnerName: "demo", Name: "Test Room", ModelName: "model_a", TradeMode: roommodel.TradeMode(9)}, expected: ErrInvalidTradeMode},
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			_, err := New(newFakeStore(), fakeLayouts{found: true, enabled: true}).Create(context.Background(), test.params)
			if !errors.Is(err, test.expected) {
				t.Fatalf("expected %v, got %v", test.expected, err)
			}
		})
	}
}

// newFakeStore creates a room store for tests.
func newFakeStore() *fakeStore {
	return &fakeStore{room: roommodel.Room{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 9}}, Name: "Test Room"}, found: true, deleted: true}
}

// validCreateForTest returns valid room creation input.
func validCreateForTest() CreateParams {
	return CreateParams{OwnerPlayerID: 7, OwnerName: "demo", Name: "Test Room", ModelName: "model_a"}
}

// fakeStore records room store calls for tests.
type fakeStore struct {
	// room is the returned room.
	room roommodel.Room

	// found reports whether lookups succeed.
	found bool

	// deleted reports whether delete succeeds.
	deleted bool

	// tags stores replacement tags.
	tags []string

	// limit stores the last list limit.
	limit int
}

// CreateRoom creates a room for tests.
func (store *fakeStore) CreateRoom(context.Context, repository.CreateRoomParams) (roommodel.Room, error) {
	return store.room, nil
}

// FindRoomByID finds a room for tests.
func (store *fakeStore) FindRoomByID(context.Context, int64) (roommodel.Room, bool, error) {
	return store.room, store.found, nil
}

// ListRoomsByOwner lists owner rooms for tests.
func (store *fakeStore) ListRoomsByOwner(context.Context, int64) ([]roommodel.Room, error) {
	return []roommodel.Room{store.room}, nil
}

// ListPopularRooms lists popular rooms for tests.
func (store *fakeStore) ListPopularRooms(_ context.Context, limit int) ([]roommodel.Room, error) {
	store.limit = limit
	return []roommodel.Room{store.room}, nil
}

// ListHighestScoreRooms lists highest score rooms for tests.
func (store *fakeStore) ListHighestScoreRooms(_ context.Context, limit int) ([]roommodel.Room, error) {
	store.limit = limit
	return []roommodel.Room{store.room}, nil
}

// SoftDeleteRoom soft deletes a room for tests.
func (store *fakeStore) SoftDeleteRoom(context.Context, int64) (bool, error) {
	return store.deleted, nil
}

// ListCategories lists categories for tests.
func (store *fakeStore) ListCategories(context.Context) ([]roommodel.Category, error) {
	return []roommodel.Category{{Caption: "Social"}}, nil
}

// ListRoomTags lists room tags for tests.
func (store *fakeStore) ListRoomTags(context.Context, int64) ([]roommodel.Tag, error) {
	return nil, nil
}

// ReplaceRoomTags replaces room tags for tests.
func (store *fakeStore) ReplaceRoomTags(_ context.Context, _ int64, tags []string) error {
	store.tags = tags
	return nil
}

// fakeLayouts resolves room layouts for tests.
type fakeLayouts struct {
	// found reports whether lookups succeed.
	found bool

	// enabled reports whether the layout is enabled.
	enabled bool
}

// Create creates a layout for tests.
func (layouts fakeLayouts) Create(context.Context, layout.SaveParams) (layout.Layout, error) {
	return layout.Layout{}, nil
}

// Update updates a layout for tests.
func (layouts fakeLayouts) Update(context.Context, int64, layout.SaveParams) (layout.Layout, error) {
	return layout.Layout{}, nil
}

// FindByID finds a layout by id for tests.
func (layouts fakeLayouts) FindByID(context.Context, int64) (layout.Layout, bool, error) {
	return layout.Layout{}, false, nil
}

// FindByName finds a layout by name for tests.
func (layouts fakeLayouts) FindByName(context.Context, string) (layout.Layout, bool, error) {
	return layout.Layout{Enabled: layouts.enabled}, layouts.found, nil
}

// List lists layouts for tests.
func (layouts fakeLayouts) List(context.Context) ([]layout.Layout, error) {
	return nil, nil
}

// Catalog loads a layout catalog for tests.
func (layouts fakeLayouts) Catalog(context.Context) (*layout.Catalog, error) {
	return nil, nil
}
