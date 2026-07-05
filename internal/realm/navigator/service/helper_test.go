package service

import (
	"context"

	navmodel "github.com/niflaot/pixels/internal/realm/navigator/model"
	"github.com/niflaot/pixels/internal/realm/navigator/repository"
)

// newFakeStore creates a navigator store for tests.
func newFakeStore() *fakeStore {
	return &fakeStore{found: true, deleted: true}
}

// fakeStore records navigator store calls for tests.
type fakeStore struct {
	// search stores the last saved search.
	search repository.SaveSearchParams

	// found reports whether lookups succeed.
	found bool

	// deleted reports whether deletes succeed.
	deleted bool
}

// AddFavorite adds a favorite for tests.
func (store *fakeStore) AddFavorite(context.Context, int64, int64) error {
	return nil
}

// RemoveFavorite removes a favorite for tests.
func (store *fakeStore) RemoveFavorite(context.Context, int64, int64) error {
	return nil
}

// ListFavoriteRoomIDs lists favorite room ids for tests.
func (store *fakeStore) ListFavoriteRoomIDs(context.Context, int64) ([]int64, error) {
	return []int64{1}, nil
}

// SaveSearch saves a search for tests.
func (store *fakeStore) SaveSearch(_ context.Context, params repository.SaveSearchParams) (navmodel.SavedSearch, error) {
	store.search = params
	return navmodel.SavedSearch{PlayerID: params.PlayerID, Code: params.Code, Filter: params.Filter}, nil
}

// DeleteSearch deletes a search for tests.
func (store *fakeStore) DeleteSearch(context.Context, int64, int64) (bool, error) {
	return store.deleted, nil
}

// ListSavedSearches lists saved searches for tests.
func (store *fakeStore) ListSavedSearches(context.Context, int64) ([]navmodel.SavedSearch, error) {
	return []navmodel.SavedSearch{{Code: "hotel_view"}}, nil
}

// UpsertPreference saves preferences for tests.
func (store *fakeStore) UpsertPreference(_ context.Context, params repository.PreferenceParams) (navmodel.Preference, error) {
	return params.Preference, nil
}

// FindPreference finds preferences for tests.
func (store *fakeStore) FindPreference(context.Context, int64) (navmodel.Preference, bool, error) {
	return navmodel.Preference{PlayerID: 7}, store.found, nil
}

// UpsertCategoryPreference saves category preferences for tests.
func (store *fakeStore) UpsertCategoryPreference(_ context.Context, params repository.CategoryPreferenceParams) (navmodel.CategoryPreference, error) {
	return params.Preference, nil
}

// ListCategoryPreferences lists category preferences for tests.
func (store *fakeStore) ListCategoryPreferences(context.Context, int64) ([]navmodel.CategoryPreference, error) {
	return []navmodel.CategoryPreference{{Code: "popular"}}, nil
}

// ListLiftedRooms lists lifted rooms for tests.
func (store *fakeStore) ListLiftedRooms(context.Context) ([]navmodel.LiftedRoom, error) {
	return []navmodel.LiftedRoom{{RoomID: 1}}, nil
}
