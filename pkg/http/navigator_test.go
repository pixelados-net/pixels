package http

import (
	"context"

	navservice "github.com/niflaot/pixels/internal/realm/navigator/core"
	navmodel "github.com/niflaot/pixels/internal/realm/navigator/record"
	sharedmodel "github.com/niflaot/pixels/pkg/model"
)

// testNavigator returns an HTTP test navigator manager.
func testNavigator() navservice.Manager { return testNavigatorManager{} }

// testNavigatorManager provides navigator data for HTTP tests.
type testNavigatorManager struct{}

// AddFavorite adds a favorite room for tests.
func (testNavigatorManager) AddFavorite(context.Context, int64, int64, int32, bool) error { return nil }

// RemoveFavorite removes a favorite room for tests.
func (testNavigatorManager) RemoveFavorite(context.Context, int64, int64) error { return nil }

// ListFavoriteRoomIDs lists favorite room ids for tests.
func (testNavigatorManager) ListFavoriteRoomIDs(context.Context, int64) ([]int64, error) {
	return nil, nil
}

// SaveSearch saves a search for tests.
func (testNavigatorManager) SaveSearch(context.Context, navservice.SaveSearchParams) (navmodel.SavedSearch, error) {
	return navmodel.SavedSearch{}, nil
}

// DeleteSearch deletes a search for tests.
func (testNavigatorManager) DeleteSearch(context.Context, int64, int64) error { return nil }

// ListSavedSearches lists saved searches for tests.
func (testNavigatorManager) ListSavedSearches(context.Context, int64) ([]navmodel.SavedSearch, error) {
	return nil, nil
}

// SavePreference saves preferences for tests.
func (testNavigatorManager) SavePreference(context.Context, navmodel.Preference) (navmodel.Preference, error) {
	return navmodel.Preference{}, nil
}

// Preference returns navigator preferences for tests.
func (testNavigatorManager) Preference(context.Context, int64) (navmodel.Preference, error) {
	return navmodel.Preference{}, nil
}

// SaveCategoryPreference saves category preferences for tests.
func (testNavigatorManager) SaveCategoryPreference(context.Context, navmodel.CategoryPreference) (navmodel.CategoryPreference, error) {
	return navmodel.CategoryPreference{}, nil
}

// ListCategoryPreferences lists category preferences for tests.
func (testNavigatorManager) ListCategoryPreferences(context.Context, int64) ([]navmodel.CategoryPreference, error) {
	return nil, nil
}

// ListLiftedRooms lists lifted rooms for tests.
func (testNavigatorManager) ListLiftedRooms(context.Context) ([]navmodel.LiftedRoom, error) {
	return []navmodel.LiftedRoom{{Base: sharedmodel.Base{Identity: sharedmodel.Identity{ID: 1}}, RoomID: 1}}, nil
}

// RecordVisit accepts one navigator visit for tests.
func (testNavigatorManager) RecordVisit(context.Context, int64, int64) error { return nil }

// ListRecentRoomIDs lists deterministic recent rooms for tests.
func (testNavigatorManager) ListRecentRoomIDs(context.Context, int64, int) ([]int64, error) {
	return []int64{1}, nil
}

// ListFrequentRoomIDs lists deterministic frequent rooms for tests.
func (testNavigatorManager) ListFrequentRoomIDs(context.Context, int64, int) ([]int64, error) {
	return []int64{1}, nil
}
