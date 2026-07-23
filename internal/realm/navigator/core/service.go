package core

import (
	"context"

	"github.com/niflaot/pixels/internal/realm/navigator/record"
)

// Manager manages Navigator persistence.
type Manager interface {
	// AddFavorite adds a favorite room for a player.
	AddFavorite(ctx context.Context, playerID int64, roomID int64, limit int32, unlimited bool) error
	// RemoveFavorite removes a favorite room for a player.
	RemoveFavorite(ctx context.Context, playerID int64, roomID int64) error
	// ListFavoriteRoomIDs lists favorite room identifiers for a player.
	ListFavoriteRoomIDs(ctx context.Context, playerID int64) ([]int64, error)
	// SaveSearch saves a Navigator search.
	SaveSearch(ctx context.Context, params SaveSearchParams) (record.SavedSearch, error)
	// DeleteSearch deletes a saved search.
	DeleteSearch(ctx context.Context, playerID int64, id int64) error
	// ListSavedSearches lists saved searches for a player.
	ListSavedSearches(ctx context.Context, playerID int64) ([]record.SavedSearch, error)
	// SavePreference saves Navigator preferences.
	SavePreference(ctx context.Context, preference record.Preference) (record.Preference, error)
	// Preference finds or returns default Navigator preferences.
	Preference(ctx context.Context, playerID int64) (record.Preference, error)
	// SaveCategoryPreference saves result-list display state.
	SaveCategoryPreference(ctx context.Context, preference record.CategoryPreference) (record.CategoryPreference, error)
	// ListCategoryPreferences lists result-list display state.
	ListCategoryPreferences(ctx context.Context, playerID int64) ([]record.CategoryPreference, error)
	// ListLiftedRooms lists currently active lifted rooms.
	ListLiftedRooms(ctx context.Context) ([]record.LiftedRoom, error)
	// RecordVisit records one admitted room visit.
	RecordVisit(ctx context.Context, playerID int64, roomID int64) error
	// ListRecentRoomIDs lists recent room history.
	ListRecentRoomIDs(ctx context.Context, playerID int64, limit int) ([]int64, error)
	// ListFrequentRoomIDs lists frequent room history.
	ListFrequentRoomIDs(ctx context.Context, playerID int64, limit int) ([]int64, error)
}

// Service validates and coordinates navigator persistence behavior.
type Service struct {
	// store reads and writes navigator persistence records.
	store record.Store
}

// New creates a navigator service.
func New(store record.Store) *Service {
	return &Service{store: store}
}

// managerAssertion verifies Service implements Manager.
var managerAssertion Manager = (*Service)(nil)
