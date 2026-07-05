package service

import (
	"context"
	"errors"
	"testing"

	navmodel "github.com/niflaot/pixels/internal/realm/navigator/model"
)

// TestFavoriteValidatesIDs verifies favorite validation.
func TestFavoriteValidatesIDs(t *testing.T) {
	err := New(newFakeStore()).AddFavorite(context.Background(), 0, 1)
	if !errors.Is(err, ErrInvalidPlayerID) {
		t.Fatalf("expected invalid player id, got %v", err)
	}
}

// TestListFavoritesRejectsInvalidPlayer verifies favorite list validation.
func TestListFavoritesRejectsInvalidPlayer(t *testing.T) {
	_, err := New(newFakeStore()).ListFavoriteRoomIDs(context.Background(), 0)
	if !errors.Is(err, ErrInvalidPlayerID) {
		t.Fatalf("expected invalid player id, got %v", err)
	}
}

// TestFavoriteListReadsStore verifies favorite listing.
func TestFavoriteListReadsStore(t *testing.T) {
	ids, err := New(newFakeStore()).ListFavoriteRoomIDs(context.Background(), 7)
	if err != nil {
		t.Fatalf("list favorites: %v", err)
	}
	if len(ids) != 1 || ids[0] != 1 {
		t.Fatalf("unexpected ids %#v", ids)
	}
}

// TestRemoveFavoriteUsesStore verifies favorite removal.
func TestRemoveFavoriteUsesStore(t *testing.T) {
	err := New(newFakeStore()).RemoveFavorite(context.Background(), 7, 9)
	if err != nil {
		t.Fatalf("remove favorite: %v", err)
	}
}

// TestSaveSearchNormalizesInput verifies saved search normalization.
func TestSaveSearchNormalizesInput(t *testing.T) {
	store := newFakeStore()
	search, err := New(store).SaveSearch(context.Background(), SaveSearchParams{PlayerID: 7, Code: " hotel_view ", Filter: " demo "})
	if err != nil {
		t.Fatalf("save search: %v", err)
	}

	if search.Code != "hotel_view" || store.search.Filter != "demo" {
		t.Fatalf("unexpected search %#v store=%#v", search, store.search)
	}
}

// TestSaveSearchRejectsInvalidInput verifies saved search validation.
func TestSaveSearchRejectsInvalidInput(t *testing.T) {
	_, err := New(newFakeStore()).SaveSearch(context.Background(), SaveSearchParams{PlayerID: 7})
	if !errors.Is(err, ErrInvalidSearch) {
		t.Fatalf("expected invalid search, got %v", err)
	}
}

// TestDeleteSearchReportsMissing verifies missing search delete behavior.
func TestDeleteSearchReportsMissing(t *testing.T) {
	store := newFakeStore()
	store.deleted = false

	err := New(store).DeleteSearch(context.Background(), 7, 8)
	if !errors.Is(err, ErrSearchNotFound) {
		t.Fatalf("expected missing search, got %v", err)
	}
}

// TestDeleteSearchRejectsInvalidID verifies saved search id validation.
func TestDeleteSearchRejectsInvalidID(t *testing.T) {
	err := New(newFakeStore()).DeleteSearch(context.Background(), 7, 0)
	if !errors.Is(err, ErrInvalidSearchID) {
		t.Fatalf("expected invalid search id, got %v", err)
	}
}

// TestListSavedSearchesReadsStore verifies saved search listing.
func TestListSavedSearchesReadsStore(t *testing.T) {
	searches, err := New(newFakeStore()).ListSavedSearches(context.Background(), 7)
	if err != nil {
		t.Fatalf("list searches: %v", err)
	}
	if len(searches) != 1 || searches[0].Code != "hotel_view" {
		t.Fatalf("unexpected searches %#v", searches)
	}
}

// TestPreferenceReturnsDefault verifies missing preferences return defaults.
func TestPreferenceReturnsDefault(t *testing.T) {
	store := newFakeStore()
	store.found = false

	preference, err := New(store).Preference(context.Background(), 7)
	if err != nil {
		t.Fatalf("load preference: %v", err)
	}

	if preference.WindowWidth != DefaultWindowWidth || preference.PlayerID != 7 {
		t.Fatalf("unexpected preference %#v", preference)
	}
}

// TestPreferenceReturnsStoredValue verifies stored preference lookup.
func TestPreferenceReturnsStoredValue(t *testing.T) {
	preference, err := New(newFakeStore()).Preference(context.Background(), 7)
	if err != nil {
		t.Fatalf("load preference: %v", err)
	}
	if preference.PlayerID != 7 {
		t.Fatalf("unexpected preference %#v", preference)
	}
}

// TestSavePreferencePersistsValidPreference verifies preference persistence.
func TestSavePreferencePersistsValidPreference(t *testing.T) {
	preference, err := New(newFakeStore()).SavePreference(context.Background(), DefaultPreference(7))
	if err != nil {
		t.Fatalf("save preference: %v", err)
	}
	if preference.PlayerID != 7 {
		t.Fatalf("unexpected preference %#v", preference)
	}
}

// TestSavePreferenceRejectsInvalidPreference verifies preference validation.
func TestSavePreferenceRejectsInvalidPreference(t *testing.T) {
	_, err := New(newFakeStore()).SavePreference(context.Background(), navmodel.Preference{PlayerID: 7})
	if !errors.Is(err, ErrInvalidPreference) {
		t.Fatalf("expected invalid preference, got %v", err)
	}
}

// TestSaveCategoryPreferenceValidatesCode verifies category preference validation.
func TestSaveCategoryPreferenceValidatesCode(t *testing.T) {
	_, err := New(newFakeStore()).SaveCategoryPreference(context.Background(), navmodel.CategoryPreference{PlayerID: 7})
	if !errors.Is(err, ErrInvalidPreference) {
		t.Fatalf("expected invalid preference, got %v", err)
	}
}

// TestSaveCategoryPreferencePersistsState verifies category preference saves.
func TestSaveCategoryPreferencePersistsState(t *testing.T) {
	preference, err := New(newFakeStore()).SaveCategoryPreference(context.Background(), navmodel.CategoryPreference{PlayerID: 7, Code: " popular ", Collapsed: true})
	if err != nil {
		t.Fatalf("save category preference: %v", err)
	}
	if preference.Code != "popular" || !preference.Collapsed {
		t.Fatalf("unexpected preference %#v", preference)
	}
}

// TestListLiftedRoomsReadsStore verifies lifted room listing.
func TestListLiftedRoomsReadsStore(t *testing.T) {
	rooms, err := New(newFakeStore()).ListLiftedRooms(context.Background())
	if err != nil {
		t.Fatalf("list lifted: %v", err)
	}
	if len(rooms) != 1 || rooms[0].RoomID != 1 {
		t.Fatalf("unexpected lifted rooms %#v", rooms)
	}
}

// TestListCategoryPreferencesReadsStore verifies category preference listing.
func TestListCategoryPreferencesReadsStore(t *testing.T) {
	preferences, err := New(newFakeStore()).ListCategoryPreferences(context.Background(), 7)
	if err != nil {
		t.Fatalf("list category preferences: %v", err)
	}
	if len(preferences) != 1 || preferences[0].Code != "popular" {
		t.Fatalf("unexpected preferences %#v", preferences)
	}
}
