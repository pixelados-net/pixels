package tests

import (
	"context"
	"testing"

	"github.com/niflaot/pixels/internal/realm/room/layout"
	roommodel "github.com/niflaot/pixels/internal/realm/room/model"
	"github.com/niflaot/pixels/internal/realm/room/repository"
	roomservice "github.com/niflaot/pixels/internal/realm/room/service"
)

// TestSearchNormalizesLimit verifies room search limit normalization.
func TestSearchNormalizesLimit(t *testing.T) {
	store := &fakeStore{}
	_, err := roomservice.New(store, fakeLayouts{}).Search(context.Background(), " demo ", 200)
	if err != nil {
		t.Fatalf("search rooms: %v", err)
	}
	if store.limit != 100 {
		t.Fatalf("expected capped limit, got %d", store.limit)
	}
}

// fakeStore records room store calls for tests.
type fakeStore struct {
	// limit stores the last list limit.
	limit int
}

// CreateRoom creates a room for tests.
func (store *fakeStore) CreateRoom(context.Context, repository.CreateRoomParams) (roommodel.Room, error) {
	return roommodel.Room{}, nil
}

// UpdateRoom updates a room for tests.
func (store *fakeStore) UpdateRoom(context.Context, repository.UpdateRoomParams, []string) (roommodel.Room, bool, error) {
	return roommodel.Room{}, false, nil
}

// FindRoomByID finds a room for tests.
func (store *fakeStore) FindRoomByID(context.Context, int64) (roommodel.Room, bool, error) {
	return roommodel.Room{}, false, nil
}

// ListRoomsByOwner lists owner rooms for tests.
func (store *fakeStore) ListRoomsByOwner(context.Context, int64) ([]roommodel.Room, error) {
	return nil, nil
}

// ListPopularRooms lists popular rooms for tests.
func (store *fakeStore) ListPopularRooms(_ context.Context, limit int) ([]roommodel.Room, error) {
	store.limit = limit
	return nil, nil
}

// ListHighestScoreRooms lists highest score rooms for tests.
func (store *fakeStore) ListHighestScoreRooms(_ context.Context, limit int) ([]roommodel.Room, error) {
	store.limit = limit
	return nil, nil
}

// SearchRooms searches rooms for tests.
func (store *fakeStore) SearchRooms(_ context.Context, _ string, limit int) ([]roommodel.Room, error) {
	store.limit = limit
	return nil, nil
}

// SoftDeleteRoom soft deletes a room for tests.
func (store *fakeStore) SoftDeleteRoom(context.Context, int64) (bool, error) {
	return false, nil
}

// ListCategories lists categories for tests.
func (store *fakeStore) ListCategories(context.Context) ([]roommodel.Category, error) {
	return nil, nil
}

// ListRoomTags lists room tags for tests.
func (store *fakeStore) ListRoomTags(context.Context, int64) ([]roommodel.Tag, error) {
	return nil, nil
}

// ReplaceRoomTags replaces room tags for tests.
func (store *fakeStore) ReplaceRoomTags(context.Context, int64, []string) error {
	return nil
}

// fakeLayouts resolves room layouts for tests.
type fakeLayouts struct{}

// Create creates a layout for tests.
func (fakeLayouts) Create(context.Context, layout.SaveParams) (layout.Layout, error) {
	return layout.Layout{}, nil
}

// Update updates a layout for tests.
func (fakeLayouts) Update(context.Context, int64, layout.SaveParams) (layout.Layout, error) {
	return layout.Layout{}, nil
}

// FindByID finds a layout by id for tests.
func (fakeLayouts) FindByID(context.Context, int64) (layout.Layout, bool, error) {
	return layout.Layout{}, false, nil
}

// FindByName finds a layout by name for tests.
func (fakeLayouts) FindByName(context.Context, string) (layout.Layout, bool, error) {
	return layout.Layout{}, false, nil
}

// List lists layouts for tests.
func (fakeLayouts) List(context.Context) ([]layout.Layout, error) {
	return nil, nil
}

// Catalog loads a layout catalog for tests.
func (fakeLayouts) Catalog(context.Context) (*layout.Catalog, error) {
	return nil, nil
}
