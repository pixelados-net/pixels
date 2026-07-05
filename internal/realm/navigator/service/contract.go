package service

import (
	"context"

	navmodel "github.com/niflaot/pixels/internal/realm/navigator/model"
)

// Manager manages navigator persistence.
type Manager interface {
	// AddFavorite adds a favorite room for a player.
	AddFavorite(ctx context.Context, playerID int64, roomID int64) error

	// RemoveFavorite removes a favorite room for a player.
	RemoveFavorite(ctx context.Context, playerID int64, roomID int64) error

	// ListFavoriteRoomIDs lists favorite room ids for a player.
	ListFavoriteRoomIDs(ctx context.Context, playerID int64) ([]int64, error)

	// SaveSearch saves a navigator search.
	SaveSearch(ctx context.Context, params SaveSearchParams) (navmodel.SavedSearch, error)

	// DeleteSearch deletes a saved search.
	DeleteSearch(ctx context.Context, playerID int64, id int64) error

	// ListSavedSearches lists saved searches for a player.
	ListSavedSearches(ctx context.Context, playerID int64) ([]navmodel.SavedSearch, error)

	// SavePreference saves navigator preferences.
	SavePreference(ctx context.Context, preference navmodel.Preference) (navmodel.Preference, error)

	// Preference finds or returns default navigator preferences.
	Preference(ctx context.Context, playerID int64) (navmodel.Preference, error)

	// SaveCategoryPreference saves result-list display state.
	SaveCategoryPreference(ctx context.Context, preference navmodel.CategoryPreference) (navmodel.CategoryPreference, error)

	// ListCategoryPreferences lists result-list display state.
	ListCategoryPreferences(ctx context.Context, playerID int64) ([]navmodel.CategoryPreference, error)

	// ListLiftedRooms lists currently active lifted rooms.
	ListLiftedRooms(ctx context.Context) ([]navmodel.LiftedRoom, error)
}

// managerAssertion verifies Service implements Manager.
var managerAssertion Manager = (*Service)(nil)
